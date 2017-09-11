package bot

import (
	"regexp"
	"github.com/bwmarrin/discordgo"
	"fmt"
	"strings"
)

var targetsMessage = regexp.MustCompile("^!(\\d+)")
var commands []func(*discordgo.Session, *discordgo.MessageCreate) bool

func init() {
	cmds := make(([]func(*discordgo.Session, *discordgo.MessageCreate) bool), 3)
	cmds[0] = bugReport
	cmds[1] = listenChannel
	cmds[2] = forgetChannel
	commands = cmds
}

func processCommand(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	for _, c := range commands {
		if c(s, m) {
			return true
		}
	}
	return false
}

func bugReport(s *discordgo.Session, m *discordgo.MessageCreate) (bool) {
	// ignore if bot
	if s.State.User.Bot {
		return false
	}

	// ignore other users
	if s.State.User.ID != m.Author.ID {
		return false
	}

	if !listeningToChannel(m.ChannelID) {
		return false
	}

	groups := targetsMessage.FindStringSubmatch(m.Content)
	if len(groups) == 2 {
		id := groups[1]
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		if msg, err := s.State.Message(m.ChannelID, id); err == nil {
			processMessage(s, msg)
			return true
		}
	}

	return false
}

func listenChannel(s *discordgo.Session, m *discordgo.MessageCreate) (bool) {
	if len(m.Mentions) != 1 || m.Mentions[0].ID != s.State.User.ID {
		return false
	}

	if strings.HasPrefix(m.Content, "!listen") {
		if ch, err := s.Channel(m.ChannelID); err == nil {
			auto := strings.HasPrefix(m.Content, "!listen auto")
			fmt.Println("Listening to channel", ch.ID, "auto =", auto)
			AddChannel(ch.ID, auto)
			s.ChannelMessageDelete(ch.ID, m.ID)
			s.State.ChannelAdd(ch)
		}
		return true
	}

	return false
}

func forgetChannel(s *discordgo.Session, m *discordgo.MessageCreate) (bool) {
	if len(m.Mentions) != 1 || m.Mentions[0].ID != s.State.User.ID {
		return false
	}

	if strings.HasPrefix(m.Content, "!forget") {
		if ch, err := s.Channel(m.ChannelID); err == nil {
			fmt.Println("Forgetting channel", ch.ID)
			RemoveChannel(ch.ID)
			s.ChannelMessageDelete(ch.ID, m.ID)
			s.State.ChannelRemove(ch)
		}
		return true
	}

	return false
}

func listeningToChannel(id string) bool {
	return ForEach(func(ch string, auto bool) bool {
		if ch == id {
			return true
		}
		return false
	})
}