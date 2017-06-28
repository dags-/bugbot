package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"text/template"
	"bytes"
	"math/rand"
)

var templ = "**Beep boop - common problem detected!**\n\n**Error:**\n```{{.Error}}```\n\n**Solution:**\n{{range .Lines}}{{.}}\n{{end}}"
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

	content := strings.ToLower(m.Content)
	name := s.State.User.Username
	if strings.Contains(content, "thank") && strings.Contains(content, name) && rand.Intn(100) > 95 {
		s.ChannelMessageSend(m.ChannelID, ":regional_indicator_b::regional_indicator_e::regional_indicator_e::regional_indicator_p: :robot: :regional_indicator_b::regional_indicator_o::regional_indicator_o::regional_indicator_p:")
	}

	result, err := scan(m)
	if err == nil {
		buf := bytes.Buffer{}
		err := message.Execute(&buf, result)
		if err == nil {
			s.ChannelMessageSend(m.ChannelID, buf.String())
		}
	}
}