package bot

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
)

func onReady(s *discordgo.Session, m *discordgo.Ready) {
	fmt.Println("Bot ready!")
	if s.State.User.Bot {
		fmt.Println("Setting bot status: online")
		s.UpdateStatus(0, "online")
	}
}