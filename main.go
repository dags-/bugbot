package main

import (
	"fmt"
	"github.com/dags-/bugbot/bot"
	"flag"
	"bufio"
	"os"
	"github.com/dags-/bugbot/visionapi"
)

func main() {
	token := flag.String("token", "", "Auth token")
	visionToken := flag.String("vision", "", "Google vision API token")
	flag.Parse()

	if *token == "" {
		fmt.Println("No token provided")
		return
	}

	if *visionToken == "" {
		fmt.Println("No vision api token provided")
	}

	go handleStop()

	vision.SetToken(*visionToken)
	bot.Start(*token)
}

func handleStop() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if scanner.Text() == "stop" {
			fmt.Println("Stopping...")
			os.Exit(0)
			break
		}
	}
}