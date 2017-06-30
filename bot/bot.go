package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
)

func Start(token, url string) {
	go pollBugs(url)

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

	sc := make(chan os.Signal, 1)
	<-sc
	dg.Close()
}