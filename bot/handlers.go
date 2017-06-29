package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"text/template"
	"bytes"
)

var templ = "{{.Mention}}\n\n**Beep boop - common problem detected!**\n\n**Error:**\n```{{.Error}}```\n\n**Solution:**\n{{range .Lines}}{{.}}\n{{end}}"
var message = template.Must(template.New("root").Parse(templ))

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