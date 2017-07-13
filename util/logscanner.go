package util

import (
	"bufio"
	"io"
)

type LogScanner struct {
	scanner *bufio.Scanner
	html    bool
}

func NewLogScanner(reader io.Reader, html bool) *LogScanner {
	return &LogScanner{
		scanner: bufio.NewScanner(reader),
		html: html,
	}
}

func (s *LogScanner) Scan() bool {
	return s.scanner.Scan()
}

func (s *LogScanner) Text() string {
	if s.html {
		return StripTags(s.scanner.Text())
	}
	return s.scanner.Text()
}
