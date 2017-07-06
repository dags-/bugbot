package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"github.com/dags-/bugbot/issue"
)

const teacher = "bugbot-teacher"
const channel = "support"

func Start(token string) {
	go issue.Init()

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
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