package issue

type Issue struct {
	Match       string `json:"error"`
	Description []string `json:"lines"`
}

func Init() {
	load()
}
