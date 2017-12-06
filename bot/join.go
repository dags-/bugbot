package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

func onJoin(s *discordgo.Session, g *discordgo.GuildCreate) {
	fmt.Println("Joined guild:", g.Name)
}