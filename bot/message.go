package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"bufio"
	"sync"
	"github.com/mvdan/xurls"
	"net/http"
	"fmt"
	"net/url"
	"io/ioutil"
	"math/rand"
	"regexp"
	"encoding/json"
)

var traceMatcher = regexp.MustCompile("(.*?Exception.+?[\n])?(\\sat (.+)[:])")

const beepboop = ":regional_indicator_b::regional_indicator_e::regional_indicator_e::regional_indicator_p: :robot: :regional_indicator_b::regional_indicator_o::regional_indicator_o::regional_indicator_p:"

type Bug struct {
	Error string `json:"error"`
	Lines []string `json:"lines"`
}

type Response struct {
	Title  string
	Error  string
	Source string
	Lines  []string
}

type Result struct {
	Mention   string
	Responses []Response
}

type GithubSearch struct {
	Total int `json:"total_count"`
}

func handleReactions(s *discordgo.Session, m *discordgo.MessageCreate) (string, bool) {
	name := strings.ToLower(s.State.User.Username)
	text := strings.ToLower(m.Content)
	if strings.Contains(text, "thank") && strings.Contains(text, name) && rand.Intn(20) >= 15 {
		return beepboop, true
	}
	return "", false
}

func handleMessage(m *discordgo.MessageCreate) (Result, bool) {
	mainFin := make(chan bool)
	mainCh := make(chan Response)

	searchFin := make(chan bool)
	searchCh := make(chan Response)

	go handleContent(m, true, mainCh, mainFin)
	go handleURLS(m, true, mainCh, mainFin)
	go handleAttachments(m, true, mainCh, mainFin)
	go handleContent(m, false, searchCh, searchFin)
	go handleURLS(m, false, searchCh, searchFin)
	go handleAttachments(m, false, searchCh, searchFin)

	if result, ok := handleResponses(mainCh, mainFin, 3); ok {
		close(searchFin)
		close(searchCh)
		return result, true
	}

	return handleResponses(searchCh, searchFin, 3)
}

func handleResponses(ch chan Response, fin chan bool, count int) (Result, bool) {
	var result Result

	for {
		select {
		case resp := <-ch:
			result.Responses = append(result.Responses, resp)
			return result, true
		case <-fin:
			count--
			if count <= 0 {
				return result, false
			}
		}
	}

	return result, false
}

func handleContent(m *discordgo.MessageCreate, known bool, ch chan Response, fin chan bool) {
	if known {
		reader := strings.NewReader(m.Content)
		scanner := bufio.NewScanner(reader)
		scanOne(scanner, "message", false, ch)
	} else {
		stackTrace(m.Content, "message", ch)
	}
	fin <- true
}

func handleURLS(m *discordgo.MessageCreate, known bool, ch chan Response, quit chan bool) {
	wg := sync.WaitGroup{}
	urls := xurls.Relaxed.FindAllString(m.Content, -1)
	for _, u := range urls {
		wg.Add(1)
		if known {
			go scanURL(u, u, true, ch, &wg)
		} else {
			go urlTrace(u, u, true, ch, &wg)
		}
	}
	wg.Wait()
	quit <- true
}

func handleAttachments(m *discordgo.MessageCreate, known bool, ch chan Response, quit chan bool) {
	wg := sync.WaitGroup{}
	for _, attachment := range m.Attachments {
		file := attachment.Filename
		if strings.HasSuffix(file, ".txt") || strings.HasSuffix(file, ".log") || strings.HasSuffix(file, ".json") {
			wg.Add(1)
			if known {
				go scanURL(attachment.URL, file, false, ch, &wg)
			} else {
				go urlTrace(attachment.URL, file, false, ch, &wg)
			}
		}
	}
	wg.Wait()
	quit <- true
}

func scanURL(url, source string, parseHtml bool, ch chan Response, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err == nil {
		scanner := bufio.NewScanner(resp.Body)
		scanOne(scanner, source, parseHtml, ch)
	}
}

func scanOne(scanner *bufio.Scanner, source string, parseHtml bool, ch chan Response) {
	for scanner.Scan() {
		l := scanner.Text()
		if parseHtml {
			l = StripTags(l)
		}
		checkOne(l, source, ch)
	}
}

func checkOne(line, source string, ch chan Response) {
	bugs := getBugs()
	for _, bug := range bugs {
		if strings.Contains(line, bug.Error) {
			ch <- Response{
				Title: "common problem detected!",
				Error: line,
				Source: source,
				Lines: bug.Lines,
			}
		}
	}
}

func urlTrace(url, src string, escape bool, ch chan Response, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	content := string(all)
	if escape {
		content = StripTags(content)
	}

	stackTrace(content, src, ch)
}

func stackTrace(content, src string, ch chan Response) {
	// number of lines to match in the stacktrace to search - more lines == more specific search .: less results
	lines := 3

	var trace []string

	matches := traceMatcher.FindAllStringSubmatch(content, lines)
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

	ch <- Response{
		Title: title,
		Error: line,
		Source: src,
		Lines: description,
	}
}

func getDescription(address string, total int) ([]string) {
	if total == 0 {
		return []string {
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

	return []string {
		"Sorry, I have not learnt about this error yet :[",
		second,
		"You may be able to find a solution here:",
		"",
		address,
	}
}