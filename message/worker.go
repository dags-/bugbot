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

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			scanner := util.NewLogScanner(strings.NewReader(m.Content), false)
			processScanner(worker, scanner, "common error detected!", "message")
		}()

		go func() {
			defer wg.Done()
			scanner := util.NewLogScanner(strings.NewReader(m.Content), false)
			lookupScanner(worker, scanner, "message")
		}()

		wg.Wait()
	}()

	return worker
}

func urlWorker(done chan interface{}, m *Message) (*worker) {
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
				processURL(worker, u, true, "common error detected!", u)
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

func attachmentWorker(done chan interface{}, m *Message) (*worker) {
	worker := newWorker(done)

	go func() {
		defer close(worker.lookups)
		defer close(worker.results)

		resources := m.Resources
		wg := sync.WaitGroup{}
		wg.Add(len(resources) * 2)

		for _, r := range resources {
			go func() {
				defer wg.Done()
				processURL(worker, r.URL, false, "common error detected!", r.Name)
			}()

			go func() {
				defer wg.Done()
				lookupURL(worker, r.URL, false, r.Name)
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
		wg.Add(len(thumbnails) * 2)

		for _, t := range thumbnails {
			go func() {
				defer wg.Done()
				processImage(worker, t.URL, "common error detected!", t.Name)
			}()

			go func() {
				defer wg.Done()
			}()
		}

		wg.Wait()
	}()

	return worker
}