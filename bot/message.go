package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"bufio"
	"sync"
	"github.com/mvdan/xurls"
	"net/http"
	"fmt"
	"net/url"
	"io/ioutil"
	"math/rand"
	"regexp"
)

var stackRegexp = regexp.MustCompile(".*?Exception.+?[\n](\\s(at.+)[\n])")

const beepboop = ":regional_indicator_b::regional_indicator_e::regional_indicator_e::regional_indicator_p: :robot: :regional_indicator_b::regional_indicator_o::regional_indicator_o::regional_indicator_p:"

type Bug struct {
	Error string `json:"error"`
	Lines []string `json:"lines"`
}

type Response struct {
	Title  string
	Error  string
	Source string
	Lines  []string
}

type Result struct {
	Mention   string
	Responses []Response
}

func handleReactions(s *discordgo.Session, m *discordgo.MessageCreate) (string, bool) {
	name := strings.ToLower(s.State.User.Username)
	text := strings.ToLower(m.Content)
	if strings.Contains(text, "thank") && strings.Contains(text, name) && rand.Intn(20) >= 15 {
		return beepboop, true
	}
	return "", false
}

func handleMessage(m *discordgo.MessageCreate) (Result, bool) {
	fin := make(chan bool)
	ch := make(chan Response)

	go handleContent(m, true, ch, fin)
	go handleURLS(m, true, ch, fin)
	go handleAttachments(m, true, ch, fin)

	result, ok := handleResponses(ch, fin, 3)
	if ok {
		return result, true
	}

	// slower
	go handleContent(m, false, ch, fin)
	go handleURLS(m, false, ch, fin)
	go handleAttachments(m, false, ch, fin)

	return handleResponses(ch, fin, 3)
}

func handleResponses(ch chan Response, fin chan bool, count int) (Result, bool) {
	var result Result

	for {
		select {
		case resp := <-ch:
			result.Responses = append(result.Responses, resp)
			return result, true
		case <-fin:
			count--
			if count <= 0 {
				return result, false
			}
		}
	}

	return result, false
}

func handleContent(m *discordgo.MessageCreate, known bool, ch chan Response, fin chan bool) {
	if known {
		reader := strings.NewReader(m.Content)
		scanner := bufio.NewScanner(reader)
		scanOne(scanner, "message", false, ch)
	} else {
		stackTrace(m.Content, "message", ch)
	}
	fin <- true
}

func handleURLS(m *discordgo.MessageCreate, known bool, ch chan Response, quit chan bool) {
	wg := sync.WaitGroup{}
	urls := xurls.Relaxed.FindAllString(m.Content, -1)
	for _, u := range urls {
		wg.Add(1)
		if known {
			go scanURL(u, u, true, ch, &wg)
		} else {
			go urlTrace(u, u, true, ch, &wg)
		}
	}
	wg.Wait()
	quit <- true
}

func handleAttachments(m *discordgo.MessageCreate, known bool, ch chan Response, quit chan bool) {
	wg := sync.WaitGroup{}
	for _, attachment := range m.Attachments {
		file := attachment.Filename
		if strings.HasSuffix(file, ".txt") || strings.HasSuffix(file, ".log") || strings.HasSuffix(file, ".json") {
			wg.Add(1)
			if known {
				go scanURL(attachment.URL, file, false, ch, &wg)
			} else {
				go urlTrace(attachment.URL, file, false, ch, &wg)
			}
		}
	}
	wg.Wait()
	quit <- true
}

func scanURL(url, source string, parseHtml bool, ch chan Response, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err == nil {
		scanner := bufio.NewScanner(resp.Body)
		scanOne(scanner, source, parseHtml, ch)
	}
}

func scanOne(scanner *bufio.Scanner, source string, parseHtml bool, ch chan Response) {
	for scanner.Scan() {
		l := scanner.Text()
		if parseHtml {
			l = StripTags(l)
		}
		checkOne(l, source, ch)
	}
}

func checkOne(line, source string, ch chan Response) {
	bugs := getBugs()
	for _, bug := range bugs {
		if strings.Contains(line, bug.Error) {
			ch <- Response{
				Title: "common problem detected!",
				Error: line,
				Source: source,
				Lines: bug.Lines,
			}
		}
	}
}

func urlTrace(url, src string, escape bool, ch chan Response, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	content := string(all)
	if escape {
		content = StripTags(content)
	}

	stackTrace(content, src, ch)
}

func stackTrace(content, src string, ch chan Response) {
	matches := stackRegexp.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		grps := matches[0]
		if len(grps) >= 3 {
			query := grps[2]
			search := fmt.Sprintf("https://google.com?#q=%s", url.QueryEscape(query))
			ch <- Response{
				Title: "unknown error!",
				Error: query,
				Source: src,
				Lines: []string{
					"I have not learnt about this error yet :[",
					"It might have been reported elsewhere online:",
					"",
					search,
				},
			}
		}
	}
}