package issue

import (
	"strings"
	"bufio"
)

func ParseMD(content string) (string) {
	var issue Issue
	var action string

	reader := strings.NewReader(content)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		switch strings.ToLower(line) {
		case "learn":
			action = "learn"
			break
		case "forget":
			action = "forget"
			break
		case "```":
			var match []string
			for scanner.Scan() {
				next := scanner.Text()
				if next == "```" {
					issue.Match = strings.Join(match, " ")
					break
				}
				match = append(match, next)
			}

			for scanner.Scan() {
				next := scanner.Text()
				issue.Description = append(issue.Description, next)
			}
			break
		}
	}

	if issue.Match != "" {
		switch action {
		case "learn":
			if len(issue.Description) > 0 {
				Learn(issue)
				return "Ok, learnt it!"
			}
			break
		case "forget":
			Forget(issue.Match)
			return "Forget what?"
		}
	}

	return ""
}