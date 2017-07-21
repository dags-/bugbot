package message

import (
	"sync"
	"net/http"
	"github.com/dags-/bugbot/util"
	"github.com/dags-/bugbot/issue"
	"golang.org/x/text/search"
	"golang.org/x/text/language"
	"github.com/dags-/bugbot/visionapi"
	"strings"
)

var lineMatcher = search.New(language.English, search.IgnoreCase)

func Process(m *Message) (Result, bool) {
	done := make(chan interface{})
	defer close(done)

	w1 := contentWorker(done, m)
	w2 := urlWorker(done, m)
	w3 := attachmentWorker(done, m)
	w4 := embedWorker(done, m)

	results := merge(done, w1.results, w2.results, w3.results, w4.results)
	lookups := merge(done, w1.lookups, w2.lookups, w3.lookups, w4.lookups)

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

	scanner := util.NewLogScanner(resp.Body, stripTags)
	processScanner(worker, scanner, title, source)
}

func processImage(worker *worker, url, title, source string) {
	query := vision.NewQuery(url, vision.TEXT, 1)
	result := vision.Post(query)
	if len(result.Responses) > 0 {
		response := result.Responses[0]
		if len(response.Annotations) > 0 {
			annotation := response.Annotations[0]
			reader := strings.NewReader(annotation.Description)
			scanner := util.NewLogScanner(reader, false)
			processScanner(worker, scanner, title, source)
		}
	}
}

func processScanner(worker *worker, scanner *util.LogScanner, title, source string) {
	for scanner.Scan() {
		text := scanner.Text()
		if processLine(worker, text, title, source) {
			return
		}
	}
}

func processLine(worker *worker, line string, title, source string) (bool) {
	result := issue.ForEach(func(issue issue.Issue) (bool) {
		if contains(line, issue.Match) {
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

func contains(s1, sub1 string) (bool) {
	index, _ := lineMatcher.IndexString(s1, sub1)
	return index != -1
}