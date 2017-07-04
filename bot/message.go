package bot

import (
	"github.com/bwmarrin/discordgo"
	"sync"
	"strings"
	"bufio"
	"github.com/mvdan/xurls"
	"net/http"
	"regexp"
	"net/url"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

var stackMatcher = regexp.MustCompile("(.*?Exception.+?[\n])?(\\sat (.+)[:])")

func handleMessage(m *discordgo.MessageCreate) (Result, bool) {
	done := make(chan interface{})
	defer close(done)

	w1 := contentWorker(done, m)
	w2 := urlWorker(done, m)
	w3 := attachmentWorker(done, m)

	results := merge(done, w1.results, w2.results, w3.results)
	lookups := merge(done, w1.lookups, w2.lookups, w3.lookups)

	result := newResult(m)
	lookup := newResult(m)

	for r := range results {
		result.Responses = append(result.Responses, r)
	}

	for l := range lookups {
		lookup.Responses = append(lookup.Responses, l)
	}

	if len(result.Responses) > 0 {
		return result, true
	}

	if len(lookup.Responses) > 0 {
		return lookup, true
	}

	var empty Result
	return empty, false
}

func merge(done chan interface{}, in ...<- chan Response) (<- chan Response) {
	var wg sync.WaitGroup
	out := make(chan Response)

	output := func(r <- chan Response) {
		defer wg.Done()

		for n := range r {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}

	wg.Add(len(in))
	for _, i := range in {
		go output(i)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func contentWorker(done chan interface{}, m *discordgo.MessageCreate) (*Worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(strings.NewReader(m.Content))
			scan(worker, scanner, false, "common error detected!", "message")
		}()

		go func() {
			defer wg.Done()
			lookupText(worker, m.Content, "message")
		}()

		wg.Wait()
	}()

	return worker
}

func urlWorker(done chan interface{}, m *discordgo.MessageCreate) (*Worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)

		urls := xurls.Relaxed.FindAllString(m.Content, -1)
		wg := sync.WaitGroup{}
		wg.Add(len(urls) * 2)

		for _, u := range urls {
			go func() {
				defer wg.Done()
				scanURL(worker, u, true, "common error detected!", u)
			}()
			go func() {
				defer wg.Done()
				lookupURL(worker, u, true, u)
			}()
		}

		wg.Wait()
	}()

	return worker
}

func attachmentWorker(done chan interface{}, m *discordgo.MessageCreate) (*Worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)

		attachments := m.Attachments
		wg := sync.WaitGroup{}
		wg.Add(len(attachments) * 2)

		for _, a := range attachments {
			u := a.URL
			src := a.Filename
			go func() {
				defer wg.Done()
				scanURL(worker, u, false, "common error detected!", src)
			}()

			go func() {
				defer wg.Done()
				lookupURL(worker, u, false, src)
			}()
		}

		wg.Wait()
	}()

	return worker
}

func scanURL(worker *Worker, url string, stripTags bool, title, source string) {
	resp, err := http.Get(url)

	if err != nil {
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	scan(worker, scanner, stripTags, title, source)
}

func scan(worker *Worker, scanner *bufio.Scanner, parseHtml bool, title, source string) {
	bugs := getBugs()

	for scanner.Scan() {
		text := scanner.Text()
		if parseHtml {
			text = StripTags(text)
		}

		if scanLine(worker, text, bugs, title, source) {
			return
		}
	}
}

func scanLine(worker *Worker, line string, bugs []Bug, title, source string) (bool) {
	for _, bug := range bugs {
		if strings.Contains(line, bug.Error) {
			select {
			case worker.results <- Response{
				Title: title,
				Source: source,
				Error: line,
				Lines: bug.Lines,
			}:
				return true
			case <-worker.done:
				return true
			}
		}
	}
	return false
}

func lookupURL(worker *Worker, url string, stripTags bool, source string) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	text := string(data)
	if stripTags {
		text = StripTags(text)
	}

	lookupText(worker, text, source)
}

func lookupText(worker *Worker, text, source string) {
	var trace []string

	matches := stackMatcher.FindAllStringSubmatch(text, 3)
	if len(matches) > 0 {
		for _, line := range matches {
			if len(line) >= 4 {
				trace = append(trace, `"` + line[3] + `"`)
			}
		}
	}

	if len(trace) == 0 {
		return
	}

	query := url.QueryEscape(strings.Join(trace, "+"))
	address := fmt.Sprintf("https://google.com?#q=%s", query)

	title := "unkown error!"
	line := strings.Trim(trace[0], `"`) + "..."
	description := getDescription(address, 0)

	if resp, err := http.Get(fmt.Sprintf("https://api.github.com/search/issues?q=%s", query)); err == nil {
		var search GithubSearch
		err := json.NewDecoder(resp.Body).Decode(&search)

		if err == nil && search.Total > 0 {
			title = "detected similar errors online"
			address = fmt.Sprintf("https://github.com/search?type=Issues&q=%s", query)
			description = getDescription(address, search.Total)
		}
	}

	select {
	case worker.lookups <- Response{
		Title: title,
		Source: source,
		Error: line,
		Lines: description,
	}:
	case <-worker.done:
		return
	}
}

func getDescription(address string, total int) ([]string) {
	if total == 0 {
		return []string{
			"Sorry, I have not learnt about this error yet :[",
			"You might be able to find more about it online:",
			"",
			address,
		}
	}

	second := "I *have*, however, found a similar issue reported online."
	if total > 1 {
		second = "I *have*, however, found similar issues reported online."
	}

	return []string{
		"Sorry, I have not learnt about this error yet :[",
		second,
		"You may be able to find a solution here:",
		"",
		address,
	}
}