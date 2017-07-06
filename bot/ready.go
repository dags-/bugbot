package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

func onReady(s *discordgo.Session, m *discordgo.Ready) {
	fmt.Println("Bot ready!")
	s.UpdateStatus(0, "online")
}