package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"text/template"
	"bytes"
)

var message = template.Must(template.ParseFiles("bot/response.html"))

func onReady(s *discordgo.Session, m *discordgo.Ready) {
	s.UpdateStatus(0, "online")
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil || strings.ToLower(channel.Name) != "support" {
		return
	}

	s.ChannelTyping(m.ChannelID)

	if react, ok := react(s, m); ok {
		s.ChannelMessageSend(m.ChannelID, react)
	}

	if response, ok := scanMessage(m); ok {
		buf := bytes.Buffer{}
		if err := message.Execute(&buf, response); err == nil {
			s.ChannelMessageSend(m.ChannelID, buf.String())
		}
	}
}