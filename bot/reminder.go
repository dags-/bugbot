package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"fmt"
	"time"
)

func remind(s *discordgo.Session) {
	sleep := time.Duration(30) * time.Minute
	for {
		for _, guild := range s.State.Guilds {
			remindGuild(s, guild)
		}

		time.Sleep(sleep)
	}
}

func remindGuild(s *discordgo.Session, guild *discordgo.Guild) {
	channels, err := s.GuildChannels(guild.ID)
	if err == nil {
		for _, ch := range channels {
			if strings.ToLower(ch.Name) == "support" {
				remindChannel(s, ch)
			}
		}
	}
}

func remindChannel(s *discordgo.Session, ch *discordgo.Channel)  {
	if ch.Topic == "" {
		return
	}

	history, err := s.ChannelMessages(ch.ID, 20, "", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(history) == 0 {
		return
	}

	for _, msg := range history {
		if strings.Contains(msg.Content, ch.Topic) {
			return
		}
	}

	s.ChannelMessageSend(ch.ID, fmt.Sprint("**Beep Boop!**\n", ch.Topic))
}