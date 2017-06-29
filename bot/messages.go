package bot

import (
	"github.com/mvdan/xurls"
	"net/http"
	"bufio"
	"strings"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"sync"
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

func scanContent(text string, ch chan Response, quit chan int) {
	reader := strings.NewReader(text)
	scanner := bufio.NewScanner(reader)
	scan(scanner, false, ch)
	quit <- 0
}

func scanContentURLs(text string, ch chan Response, quit chan int) {
	wg := sync.WaitGroup{}
	urls := xurls.Relaxed.FindAllString(text, -1)
	for _, url := range urls {
		wg.Add(1)
		go scanURL(url, true, ch, &wg)
	}
	wg.Wait()
	quit <- 0
}

func scanAttachments(attachments []*discordgo.MessageAttachment, ch chan Response, quit chan int) {
	wg := sync.WaitGroup{}
	for _, attachment := range attachments {
		file := attachment.Filename
		if strings.HasSuffix(file, ".txt") || strings.HasSuffix(file, ".log") || strings.HasSuffix(file, ".json") {
			wg.Add(1)
			go scanURL(attachment.URL, false, ch, &wg)
		}
	}
	wg.Wait()
	quit <- 0
}

func scan(scanner *bufio.Scanner, parseHtml bool, ch chan Response) {
	for scanner.Scan() {
		l := scanner.Text()
		if parseHtml {
			l = StripTags(l)
		}
		checkLine(l, ch)
	}
}

func scanURL(url string, parseHtml bool, ch chan Response, wg *sync.WaitGroup)  {
	defer wg.Done()

	resp, err := http.Get(url)
	if err == nil {
		scanner := bufio.NewScanner(resp.Body)
		scan(scanner, parseHtml, ch)
	}
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