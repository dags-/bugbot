package message

import (
	"sync"
	"strings"
	"github.com/mvdan/xurls"
	"github.com/dags-/bugbot/util"
)

func newWorker(done chan interface{}) (*worker) {
	return &worker{
		done: done,
		results: make(chan Response),
		lookups: make(chan Response),
	}
}

func contentWorker(done chan interface{}, m *Message) (*worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)
		scanner := util.NewLogScanner(strings.NewReader(m.Content), false)
		processScanner(worker, scanner, "common error detected!", "message")
	}()

	return worker
}

func urlWorker(done chan interface{}, m *Message) (*worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)

		urls := xurls.Relaxed().FindAllString(m.Content, -1)
		wg := sync.WaitGroup{}
		wg.Add(len(urls))

		for _, u := range urls {
			go func() {
				defer wg.Done()
				processURL(worker, u, true, "common error detected!", u)
			}()
		}

		wg.Wait()
	}()

	return worker
}

func attachmentWorker(done chan interface{}, m *Message) (*worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)

		resources := m.Resources
		wg := sync.WaitGroup{}
		wg.Add(len(resources))

		for _, r := range resources {
			go func() {
				defer wg.Done()
				processURL(worker, r.URL, false, "common error detected!", r.Name)
			}()
		}

		wg.Wait()
	}()

	return worker
}

func embedWorker(done chan interface{}, m *Message) (*worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)

		thumbnails := m.Thumbnails
		wg := sync.WaitGroup{}
		wg.Add(len(thumbnails))

		for _, t := range thumbnails {
			go func() {
				defer wg.Done()
				processImage(worker, t.URL, "common error detected!", t.Name)
			}()
		}

		wg.Wait()
	}()

	return worker
}