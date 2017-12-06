package main

import (
	"os"
	"fmt"
	"flag"
	"bufio"
	"github.com/dags-/bugbot/bot"
	"github.com/dags-/bugbot/visionapi"
)

func main() {
	user := flag.Bool("user", false, "Start with user token")
	email := flag.String("email", "", "Email")
	pass := flag.String("pass", "", "Password")
	token := flag.String("token", "", "Auth token")
	visionToken := flag.String("vision", "", "Google vision API token")
	teacherID := flag.String("teacher", "", "Teacher ID")
	flag.Parse()

	if *email != "" && *pass != "" && *token == "" {
		fmt.Println("Fetching user token...")
		token, err := bot.GetUserToken(*email, *pass)
		b := true
		if err != nil {
			fmt.Println(err)
			b = false
		}
		fmt.Println(*token)
		user = &b
	}

	if *token == "" {
		fmt.Println("No token provided")
		return
	}

	if *visionToken == "" {
		fmt.Println("No vision api token provided")
	}

	go handleStop()

	vision.SetToken(*visionToken)
	bot.Start(*user, *token, *teacherID)
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