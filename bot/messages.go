package bot

import (
	"github.com/mvdan/xurls"
	"net/http"
	"bufio"
	"strings"
	"github.com/bwmarrin/discordgo"
	"math/rand"
)

type Bug struct {
	Error string `json:"error"`
	Lines []string `json:"lines"`
}

type Response struct {
	Error   string
	Mention string
	Lines   []string
}

const beepboop = ":regional_indicator_b::regional_indicator_e::regional_indicator_e::regional_indicator_p: :robot: :regional_indicator_b::regional_indicator_o::regional_indicator_o::regional_indicator_p:"

func react(s *discordgo.Session, m *discordgo.MessageCreate) (string, bool) {
	name := strings.ToLower(s.State.User.Username)
	text := strings.ToLower(m.Content)
	if strings.Contains(text, "thank") && strings.Contains(text, name) && rand.Intn(20) >= 15 {
		return beepboop, true
	}
	return "", false
}

func scanMessage(msg *discordgo.MessageCreate) (Response, bool) {
	var empty Response

	count := 3
	quit := make(chan int, count)
	ch := make(chan Response, 3)

	go scanContent(msg.Content, ch, quit)
	go scanContentURLs(msg.Content, ch, quit)
	go scanAttachments(msg.Attachments, ch, quit)

	for {
		select {
		case first := <-ch:
			first.Mention = msg.Author.Mention()
			return first, true
		case <-quit:
			count--
			if count <= 0 {
				return empty, false
			}
		}
	}

	return empty, false
}

func checkLine(line string, ch chan Response) {
	bugs := getBugs()

	for _, bug := range bugs {
		if strings.Contains(line, bug.Error) {
			ch <- Response{
				Error: line,
				Mention: "",
				Lines: bug.Lines,
			}
		}
	}
}

func scan(scanner *bufio.Scanner, ch chan Response) {
	for scanner.Scan() {
		l := scanner.Text()
		checkLine(l, ch)
	}
}

func scanContent(text string, ch chan Response, quit chan int) {
	reader := strings.NewReader(text)
	scanner := bufio.NewScanner(reader)
	scan(scanner, ch)
	quit <- 0
}

func scanContentURLs(text string, ch chan Response, quit chan int) {
	urls := xurls.Relaxed.FindAllString(text, -1)
	for _, url := range urls {
		resp, err := http.Get(url)
		if err == nil {
			scanner := bufio.NewScanner(resp.Body)
			scan(scanner, ch)
		}
	}
	quit <- 0
}

func scanAttachments(attachments []*discordgo.MessageAttachment, ch chan Response, quit chan int) {
	for _, attachment := range attachments {
		resp, err := http.Get(attachment.URL)
		if err == nil {
			scanner := bufio.NewScanner(resp.Body)
			scan(scanner, ch)
		}
	}
	quit <- 0
}