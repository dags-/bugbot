package vision

import (
	"io"
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
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

func Post(query Query) Result {
	var result Result

	data, err := json.Marshal(query)
	if err == nil {
		resp, err := http.Post(apiUrl, "json", bytes.NewReader(data))
		if err == nil {
			result = Parse(resp.Body)
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}

	return result
}

func Parse(r io.Reader) Result {
	var resp Result
	err := json.NewDecoder(r).Decode(&resp)
	if err != nil {
		fmt.Println(err)
	}
	return resp
}