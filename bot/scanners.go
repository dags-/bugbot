package bot

import (
	"github.com/mvdan/xurls"
	"net/http"
	"bufio"
	"strings"
	"errors"
	"github.com/bwmarrin/discordgo"
	"fmt"
)

type Bug struct {
	Error string `json:"error"`
	Lines []string `json:"lines"`
}

var bugs []Bug

func scan(m *discordgo.MessageCreate) (Bug, error) {
	var bug Bug
	var err error

	bug, err = scanContent(m.Content)
	if err == nil {
		return bug, err
	}

	bug, err = scanMessageURLS(m.Content)
	if err == nil {
		return bug, err
	}

	bug, err = scanAttachments(m.Attachments)
	if err == nil {
		return bug, err
	}

	return bug, err
}

func scanContent(content string) (Bug, error) {
	var empty Bug

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		bug, err := scanText(line)
		if err == nil {
			return bug, nil
		}
	}

	return empty, errors.New("No content match")
}

func scanText(content string) (Bug, error) {
	var empty Bug

	for _, bug := range bugs {
		if strings.Contains(content, bug.Error) {
			b := Bug{
				Error: content,
				Lines: bug.Lines,
			}
			return b, nil
		}
	}

	return empty, errors.New("No text match")
}

func scanMessageURLS(content string) (Bug, error) {
	var empty Bug

	urls := xurls.Relaxed.FindAllString(content, -1)
	for _, url := range urls {
		resp, err := http.Get(url)
		if err == nil {
			bug, err := scanResponse(resp)
			if err == nil {
				return bug, nil
			}
		}
	}

	return empty, errors.New("No url match")
}

func scanResponse(response *http.Response) (Bug, error) {
	var empty Bug

	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		line := scanner.Text()
		bug, err := scanText(line)
		if err == nil {
			return bug, nil
		}
	}

	return empty, errors.New("No response match")
}

func scanAttachments(attachments []*discordgo.MessageAttachment) (Bug, error) {
	var empty Bug

	for _, attachment := range attachments {
		resp, err := http.Get(attachment.URL)
		if err != nil {
			fmt.Println("attach", err)
			continue
		}

		bug, err := scanResponse(resp)
		if err == nil {
			return bug, nil
		}
	}

	return empty, errors.New("No attachment response")
}