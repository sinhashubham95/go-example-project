package business

import (
	"context"
	"github.com/sinhashubham95/go-example-project/constants"
	"github.com/sinhashubham95/go-example-project/external"
	"github.com/sinhashubham95/go-example-project/models"
	"github.com/sinhashubham95/go-example-project/utils/configs"
)

// GetMoxy is used to get the moxy response
func GetMoxy(ctx context.Context) (models.MoxyResponse, error) {
	response := models.MoxyResponse{}

	moxyConfig, err := configs.Get(constants.MoxyConfig)
	if err != nil {
		return response, err
	}

	data, err := external.GetMoxy(ctx, moxyConfig.GetString(constants.URLConfigKey))
	if err != nil {
		return response, err
	}

	response.Data = data

	return response, nil
}
