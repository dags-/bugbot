package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"github.com/dags-/bugbot/issue"
)

func Start(token, devId string) {
	go issue.Init()

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(onMessage)
	dg.AddHandler(onReady)

	fmt.Println("Bot opening connection...")

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	devID = devId
	sc := make(chan os.Signal, 1)

	<-sc
	dg.Close()
}