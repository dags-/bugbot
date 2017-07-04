package bot

import "github.com/bwmarrin/discordgo"

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

type GithubSearch struct {
	Total int `json:"total_count"`
}

type Worker struct {
	done    chan interface{}
	results chan Response
	lookups chan Response
}

func newWorker(done chan interface{}) (*Worker) {
	return &Worker{
		done: done,
		results: make(chan Response),
		lookups: make(chan Response),
	}
}

func newResult(m * discordgo.MessageCreate) (Result) {
	return Result{Mention: m.Author.Mention()}
}