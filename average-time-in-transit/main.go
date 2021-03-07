package main

import (
	"context"

	"github.com/elorusso/wonderment-tech-eval/models"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, event *models.APIGatewayPayload) (string, error) {
	return "", nil
}
