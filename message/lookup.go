package message

import (
	"encoding/json"
	"strings"
	"fmt"
	"net/url"
	"regexp"
	"net/http"
	"github.com/dags-/bugbot/util"
)

var exceptionMatcher = regexp.MustCompile(`.+?Exception:.+?`)
var traceMatcher = regexp.MustCompile(`\sat\s(.+\(.+\))`)

func lookupURL(worker *worker, url string, stripTags bool, source string) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	scanner := util.NewLogScanner(resp.Body, stripTags)
	lookupScanner(worker, scanner, source)
}

func lookupScanner(worker *worker, scanner *util.LogScanner, source string) {
	for scanner.Scan() {
		line := scanner.Text()

		if exceptionMatcher.MatchString(line) {
			lines := 3
			trace := make([]string, lines)
			trace[0] = line
			for i := 1; scanner.Scan() && i <= lines; i++ {
				line = scanner.Text()
				groups := traceMatcher.FindAllStringSubmatch(line, -1)
				if len(groups) == 0 || len(groups[0]) < 2 {
					break
				}

				if i >= lines {
					lookupStackTrace(worker, trace, source)
					return
				}

				trace[i] = groups[0][1]
			}
		}
	}
}

func lookupStackTrace(worker *worker, trace []string, source string) {
	query := url.QueryEscape(strings.Join(trace, "+"))
	address := fmt.Sprintf("https://api.github.com/search/issues?q=%s", query)

	resp, err := http.Get(address)
	if err != nil {
		return
	}

	var srch GithubSearch
	err = json.NewDecoder(resp.Body).Decode(&srch)
	if err != nil || srch.Total == 0 {
		return
	}

	title := "detected similar errors online"
	line := strings.Trim(trace[0], `"`) + "..."
	address = fmt.Sprintf("https://github.com/search?type=Issues&q=%s", query)
	description := []string{
		"Sorry, I have not learnt about this error yet :[",
		"I *have* found similar issue(s) reported online.",
		"You may be able to find a solution here:",
		"",
		address,
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
