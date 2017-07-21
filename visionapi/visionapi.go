package vision

import (
	"fmt"
)

var apiToken string

func SetToken(token string) {
	apiToken = token
}

func GetURL() string {
	return fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", apiToken)
}
