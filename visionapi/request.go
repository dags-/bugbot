package vision

const TEXT = "TEXT_DETECTION"

type Query struct {
	Requests []Request `json:"requests"`
}

type Request struct {
	Image    Image `json:"image"`
	Features []Feature `json:"features"`
}

type Image struct {
	Source Source `json:"source"`
}

type Source struct {
	URI string `json:"imageUri"`
}

type Feature struct {
	Type string `json:"type"`
	Max  int `json:"maxResults"`
}

func NewQuery(url, typ string, max int) Query {
	return Query{
		Requests: []Request{
			{
				Image: Image{
					Source: Source{
						URI: url,
					},
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