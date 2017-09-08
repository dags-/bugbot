package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"github.com/dags-/bugbot/issue"
)

const teacher = "bugbot-teacher"
const channel = "support"

func StartUser(token string) {
	start("", token)
}

func StartBot(token string) {
	start("bot ", token)
}

func start(bot, token string) {
	go issue.Init()

	s, err := discordgo.New(bot + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	if bot == "" {
		s.State.MaxMessageCount = 9999
	}

	s.AddHandler(onMessage)
	s.AddHandler(onReady)
	s.AddHandler(onJoin)
	go remind(s)

	fmt.Println("Bot opening connection...")

	err = s.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	sc := make(chan os.Signal, 1)

	<-sc
	s.Close()
}

func GetUserToken(email, password string) (*string, error) {
	s, e := discordgo.New(email, password)
	t := ""
	if e == nil {
		t = s.Token
	}
	return &t, e
}