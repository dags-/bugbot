package main

import (
	"net/http"
	"fmt"
	"encoding/json"
	"github.com/dags-/bugbot/bot"
	"flag"
)

func main() {
	token := flag.String("token", "", "Auth token")
	flag.Parse()

	if *token == "" {
		fmt.Println("No token provided")
		return
	}

	resp, err := http.Get("https://dags-.github.io/bugbot/common-errors.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	var bugs []bot.Bug
	err = json.NewDecoder(resp.Body).Decode(&bugs)

	if err != nil {
		fmt.Println(err)
		return
	}

	bot.Start(*token, bugs)
}