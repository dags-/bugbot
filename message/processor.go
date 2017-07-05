package message

import (
	"sync"
	"strings"
	"bufio"
	"net/http"
	"regexp"
	"net/url"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"github.com/dags-/bugbot/util"
	"github.com/dags-/bugbot/issue"
)

var stackMatcher = regexp.MustCompile("(.*?Exception.+?[\n])?(\\sat (.+)[:])")

func Process(m *Message) (Result, bool) {
	done := make(chan interface{})
	defer close(done)

	w1 := contentWorker(done, m)
	w2 := urlWorker(done, m)
	w3 := attachmentWorker(done, m)

	results := merge(done, w1.results, w2.results, w3.results)
	lookups := merge(done, w1.lookups, w2.lookups, w3.lookups)

	for r := range results {
		result := Result{
			Mention: m.Author,
			Responses: []Response{r},
		}
		return result, true
	}

	for l := range lookups {
		lookup := Result{
			Mention: m.Author,
			Responses: []Response{l},
		}
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


func processURL(worker *worker, url string, stripTags bool, title, source string) {
	resp, err := http.Get(url)

	if err != nil {
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	processScanner(worker, scanner, stripTags, title, source)
}

func processScanner(worker *worker, scanner *bufio.Scanner, parseHtml bool, title, source string) {
	for scanner.Scan() {
		text := scanner.Text()
		if parseHtml {
			text = util.StripTags(text)
		}

		if processLine(worker, text, title, source) {
			return
		}
	}
}

func processLine(worker *worker, line string, title, source string) (bool) {
	result := issue.ForEach(func(issue issue.Issue) (bool) {
		if strings.Contains(line, issue.Match) {
			select {
			case worker.results <- Response{
				Title: title,
				Source: source,
				Error: line,
				Lines: issue.Description,
			}:
				return true
			case <-worker.done:
				return true
			}
		}
		return false
	})
	return result
}

func lookupURL(worker *worker, url string, stripTags bool, source string) {
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
		text = util.StripTags(text)
	}

	lookupText(worker, text, source)
}

func lookupText(worker *worker, text, source string) {
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

	lookupStackTrace(worker, trace, source)
}

func lookupStackTrace(worker *worker, trace []string, source string) {
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