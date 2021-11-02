package business

import (
	"context"
	"github.com/sinhashubham95/go-example-project/external"
	"github.com/sinhashubham95/go-example-project/models"
)

// GetMoxy is used to get the moxy response
func GetMoxy(ctx context.Context) (models.MoxyResponse, error) {
	response := models.MoxyResponse{}

	data, err := external.GetMoxy(ctx)
	if err != nil {
		return response, err
	}

	response.Data = data

	return response, nil
}
