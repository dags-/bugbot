package bot

import (
	"os"
	"encoding/json"
	"sync"
	"io/ioutil"
	"fmt"
)

var filename string
var memory map[string]bool
var lock = sync.RWMutex{}

func loadChannels(id string) {
	lock.Lock()
	defer lock.Unlock()

	var data []byte
	var err error
	filename = id + "-channels.json"

	if data, err = ioutil.ReadFile(filename); os.IsNotExist(err) {
		fmt.Println("Creating new channels json...")
		memory = make(map[string]bool)
		go write()
		return
	}

	json.Unmarshal(data, &memory)
}

func write() {
	lock.Lock()
	defer lock.Unlock()
	if f, err := os.Create(filename); err == nil {
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", "  ")
		if err = encoder.Encode(&memory); err != nil {
			fmt.Println(err)
		}
	}
}

func ForEach(consumer func(string, bool) (bool)) bool {
	lock.RLock()
	defer lock.RUnlock()
	for id, auto := range memory {
		if consumer(id, auto) {
			return true
		}
	}
	return false
}

func AddChannel(id string, auto bool) {
	lock.Lock()
	defer lock.Unlock()
	memory[id] = auto
	go write()
}

func RemoveChannel(id string) {
	lock.Lock()
	defer lock.Unlock()
	size := len(memory)
	delete(memory, id)
	if len(memory) != size {
		go write()
	}
}