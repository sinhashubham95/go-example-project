package processor

import (
	"errors"
	"github.com/angel-one/go-utils"
	"net/http"
)

// ProcessMoxyResponse is used to process moxy response
func ProcessMoxyResponse(response *http.Response) (string, error) {
	if response.Body == nil {
		return "", errors.New("no response exists")
	}
	return utils.GetDataAsString(response.Body)
}
