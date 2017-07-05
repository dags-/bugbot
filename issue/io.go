package issue

import (
	"io/ioutil"
	"os"
	"fmt"
	"net/http"
	"encoding/json"
)

const filename = "issues.json"
const lookup = "https://bugbot.dags.me/common-errors.json"

func load() {
	var issues []Issue
	var data []byte
	var err error
	var mk bool

	if data, err = ioutil.ReadFile(filename); os.IsNotExist(err) {
		fmt.Println("Fetching issues from:", lookup)

		mk = true
		resp, err := http.Get(lookup)
		if err != nil {
			fmt.Println(err)
			return
		}

		if data, err = ioutil.ReadAll(resp.Body); err != nil {
			fmt.Println(err)
			return
		}
	}

	err = json.Unmarshal(data, &issues)
	if err != nil {
		fmt.Println(err)
		return
	}

	putAll(issues)

	if mk {
		write()
	}
}

func write() {
	lock.RLock()
	defer lock.RUnlock()

	if file, err := os.Create(filename); err == nil {
		var slice []Issue
		for _, val := range memory {
			slice = append(slice, val)
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err := encoder.Encode(slice)

		if err != nil {
			fmt.Println(err)
		}
	}
}