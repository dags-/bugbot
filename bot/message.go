package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dags-/bugbot/message"
	"fmt"
	"github.com/dags-/bugbot/issue"
	"strings"
	"bytes"
	"text/template"
)

var templ = template.Must(template.ParseFiles("response.html"))

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore self or bots
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	// ignore other channels
	ch, err := s.Channel(m.ChannelID)
	if err != nil || strings.ToLower(ch.Name) != channel {
		return
	}

	// user has tagged the bot and has 'bugbot-teacher' role
	if mentionsBot(s, m) && canTeach(s, m) {
		s.ChannelTyping(m.ChannelID)
		if response, ok := tryLearn(s, m); ok {
			s.ChannelMessageSend(m.ChannelID, response)
			return
		}
	}

	// process message and send response if bug detected
	s.ChannelTyping(m.ChannelID)
	if response, ok := message.Process(convertMessage(m)); ok {
		buf := bytes.Buffer{}
		if err := templ.Execute(&buf, response); err == nil {
			s.ChannelMessageSend(m.ChannelID, buf.String())
			return
		}
	}
}

func convertMessage(m *discordgo.MessageCreate) (*message.Message) {
	var msg = &message.Message{}
	msg.Author = m.Author.Mention()
	msg.Content = m.Content
	msg.Resources = make([]message.Resource, len(m.Attachments))

	for i, a := range m.Attachments {
		resource := message.Resource{Name: a.Filename, URL: a.URL}
		msg.Resources[i] = resource
		if hasImageExtn(a.Filename) {
			msg.Thumbnails = append(msg.Thumbnails, resource)
		}
	}

	for _, e := range m.Embeds {
		resource := message.Resource{Name: e.Title, URL: e.Thumbnail.URL}
		msg.Thumbnails = append(msg.Thumbnails, resource)
	}

	return msg
}

func mentionsBot(s *discordgo.Session, m *discordgo.MessageCreate) (bool) {
	for _, user := range m.Mentions {
		if user.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

func canTeach(s *discordgo.Session, m *discordgo.MessageCreate) (bool) {
	var channel *discordgo.Channel
	var member *discordgo.Member
	var err error

	if channel, err = s.Channel(m.ChannelID); err != nil {
		fmt.Println(err)
		return false
	}

	if member, err = s.GuildMember(channel.GuildID, m.Author.ID); err != nil {
		fmt.Println(err)
		return false
	}

	role := getRoleId(s, channel.GuildID)
	if role == "" {
		return false
	}

	for _, owned := range member.Roles {
		if owned == role {
			return true
		}
	}

	return false
}

func getRoleId(s *discordgo.Session, guild string) (string) {
	roles, err := s.GuildRoles(guild)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	for _, role := range roles {
		if role.Name == teacher {
			return role.ID
		}
	}

	return ""
}

func tryLearn(s *discordgo.Session, m *discordgo.MessageCreate) (string, bool) {
	response := issue.ParseMD(m.Content)
	ok := response != ""
	return response, ok
}

func hasImageExtn(name string) bool {
	s := strings.ToLower(name)
	return strings.HasSuffix(s, ".jpeg") || strings.HasSuffix(s, ".jpg") || strings.HasSuffix(s, ".png")
}