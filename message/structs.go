package message

type Message struct {
	Author    string
	Content   string
	Resources []Resource
}

type Resource struct {
	Name string
	URL  string
}

type Result struct {
	Mention   string
	Responses []Response
}

type Response struct {
	Title  string
	Error  string
	Source string
	Lines  []string
}

type GithubSearch struct {
	Total int `json:"total_count"`
}

type worker struct {
	done    chan interface{}
	results chan Response
	lookups chan Response
}