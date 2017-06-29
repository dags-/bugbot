package main

import (
	"fmt"
	"github.com/dags-/bugbot/bot"
	"flag"
	"bufio"
	"os"
)

const bugs = "https://bugbot.dags.me/common-errors.json"

func main() {
	token := flag.String("token", "", "Auth token")
	errs := flag.String("errors", bugs, "Common errors url")
	flag.Parse()

	if *token == "" {
		fmt.Println("No token provided")
		return
	}

	go handleStop()

	bot.Start(*token, errs)
}

func handleStop()  {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			fmt.Println("Stopping...")
			os.Exit(0)
			break
		}
	}
}
