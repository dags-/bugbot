package bot

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