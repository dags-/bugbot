package issue

import (
	"sync"
)

var memory = make(map[string]Issue)
var lock = sync.RWMutex{}

func ForEach(consumer func(Issue)(bool)) (bool) {
	lock.RLock()
	defer lock.RUnlock()
	for _, i := range memory {
		if consumer(i) {
			return true
		}
	}
	return false
}

func putAll(issue []Issue) {
	lock.Lock()
	defer lock.Unlock()
	for _, i := range issue {
		memory[i.Match] = i
	}
}

func Learn(issue Issue) {
	lock.Lock()
	defer lock.Unlock()
	memory[issue.Match] = issue
	go write()
}

func Forget(key string) {
	lock.Lock()
	defer lock.Unlock()
	delete(memory, key)
	go write()
}