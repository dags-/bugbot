package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"github.com/dags-/bugbot/issue"
)

const teacher = "bugbot-teacher"

func Start(user bool, token string) {
	go issue.Init()

	if !user {
		token = "Bot " + token
	}

	s, err := discordgo.New(token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	if user {
		s.State.MaxMessageCount = 9999
	} else {
		go remind(s)
	}

	s.AddHandler(onMessage)
	s.AddHandler(onReady)
	s.AddHandler(onJoin)

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