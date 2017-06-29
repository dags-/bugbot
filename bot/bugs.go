package bot

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"
)

var bugsChannel = make(chan []Bug)

func getBugs() []Bug {
	return <-bugsChannel
}

func pollBugs(url string) {
	t := time.Now()
	d := time.Duration(5) * time.Minute

	bugs, _ := fetchBugs(url)

	for {
		if time.Now().Sub(t) > d {
			t = time.Now()
			if b, ok := fetchBugs(url); ok {
				bugs = b
			}
			bugsChannel <- bugs
		} else {
			bugsChannel <- bugs
		}
	}
}

func fetchBugs(url string) ([]Bug, bool) {
	fmt.Println("Fetching bugs json...")

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return make([]Bug, 0), false
	}

	var bugs []Bug
	err = json.NewDecoder(resp.Body).Decode(&bugs)

	if err != nil {
		fmt.Println(err)
		return make([]Bug, 0), false
	}

	return bugs, true
}