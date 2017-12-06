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
		time.Sleep(sleep)

		for _, guild := range s.State.Guilds {
			remindGuild(s, guild)
		}
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

	history, err := s.ChannelMessages(ch.ID, 30, "", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(history) < 30 {
		return
	}

	for _, msg := range history {
		if strings.Contains(msg.Content, ch.Topic) {
			return
		}
	}

	s.ChannelMessageSend(ch.ID, fmt.Sprint("**Beep Boop!**\n", ch.Topic))
}