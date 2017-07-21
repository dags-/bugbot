package vision

import (
	"io"
	"encoding/json"
	"fmt"
)

type Result struct {
	Responses []Response `json:"responses"`
}

type Response struct {
	Annotations []Annotation `json:"textAnnotations"`
}

type Annotation struct {
	Description string `json:"description"`
}

func Parse(r io.Reader) Result {
	var resp Result
	err := json.NewDecoder(r).Decode(&resp)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}