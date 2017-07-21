package vision

import (
	"fmt"
)

var apiUrl string

func SetToken(token string) {
	apiUrl = fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", token)
}
