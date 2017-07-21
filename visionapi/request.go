package vision

import (
	"net/http"
	"encoding/base64"
	"io/ioutil"
)

const TEXT = "TEXT_DETECTION"

type Query struct {
	Requests []Request `json:"requests"`
}

type Request struct {
	Image    Image `json:"image"`
	Features []Feature `json:"features"`
}

type Image struct {
	Content string `json:"content"`
	Source  Source `json:"source"`
}

type Source struct {
	URI string `json:"imageUri"`
}

type Feature struct {
	Type string `json:"type"`
	Max  int `json:"maxResults"`
}

func NewQuery(url, typ string, max int) Query {
	req, _ := http.Get(url)
	b, _ := ioutil.ReadAll(req.Body)
	s := base64.URLEncoding.EncodeToString(b)

	return Query{
		Requests: []Request{
			{
				Image: Image{
					Content: s,
				},
				Features: []Feature{
					{
						Type: typ,
						Max: max,
					},
				},
			},
		},
	}
}