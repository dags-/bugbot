package main

import (
	"fmt"
	"github.com/dags-/bugbot/bot"
	"flag"
	"bufio"
	"os"
)

func main() {
	token := flag.String("token", "", "Auth token")
	devId := flag.String("dev", "dags#8913", "Developer id")
	flag.Parse()

	if *token == "" {
		fmt.Println("No token provided")
		return
	}

	go handleStop()

	bot.Start(*token, *devId)
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