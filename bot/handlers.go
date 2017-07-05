package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"text/template"
	"fmt"
	"bytes"
	"github.com/dags-/bugbot/message"
	"github.com/dags-/bugbot/issue"
)

var devID string
var templ = template.Must(template.ParseFiles("response.html"))

func onReady(s *discordgo.Session, m *discordgo.Ready) {
	fmt.Println("Bot ready!")
	s.UpdateStatus(0, "online")
	go remind(s)
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil || strings.ToLower(channel.Name) != "support" {
		return
	}

	if response, ok := tryLearn(s, m); ok {
		s.ChannelMessageSend(m.ChannelID, response)
		return
	}

	s.ChannelTyping(m.ChannelID)

	if response, ok := message.Process(convert(m)); ok {
		buf := bytes.Buffer{}
		if err := templ.Execute(&buf, response); err == nil {
			s.ChannelMessageSend(m.ChannelID, buf.String())
			return
		} else {
			fmt.Println(err)
		}
	}
}

func convert(m *discordgo.MessageCreate) *message.Message {
	var msg = &message.Message{}
	msg.Author = m.Author.Mention()
	msg.Content = m.Content
	msg.Resources = make([]message.Resource, len(m.Attachments))
	for i, a := range m.Attachments {
		msg.Resources[i] = message.Resource{Name: a.Filename, URL: a.URL}
	}
	return msg
}

func tryLearn(s *discordgo.Session, m *discordgo.MessageCreate) (string, bool) {
	if m.Author.String() == devID {
		for _, user := range m.Mentions {
			if user.ID == s.State.User.ID {
				s.ChannelTyping(m.ChannelID)
				return issue.ParseMD(m.Content), true
			}
		}
	}
	return "", false
}