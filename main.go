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

	t := "MzI5OTE0MzQ0OTAzMTQ3NTIw.DDaYRw.xIc3HD7QWHuTp4UXWKpf29yV_BE"
	token = &t

	if *token == "" {
		fmt.Println("No token provided")
		return
	}

	go handleStop()

	bot.Start(*token, *errs)
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
