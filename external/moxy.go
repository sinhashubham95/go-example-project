package external

import (
	"context"
	"github.com/sinhashubham95/go-example-project/external/processor"
	"github.com/sinhashubham95/go-example-project/utils/httpclient"
)

// GetMoxy is used to get the response from the moxy service
func GetMoxy(ctx context.Context) (string, error) {
	response, err := httpclient.Get().Request(httpclient.NewRequest("moxy").SetContext(ctx))
	if err != nil {
		return "", err
	}
	return processor.ProcessMoxyResponse(response)
}
