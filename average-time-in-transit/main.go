package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	dataAccess "github.com/elorusso/wonderment-tech-eval/data-access"
	"github.com/elorusso/wonderment-tech-eval/models"
)

func main() {
	// lambda
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, payload *models.APIGatewayPayload) (*models.APIGatewayResponse, error) {
	startTime := time.Now()

	var carrier string
	//check for carrier parameter
	if payload.QueryStringParameters != nil {
		//check query params
		if carrierVal, ok := payload.QueryStringParameters["carrier"]; ok {
			carrier = carrierVal
		}
	}

	//connect to db
	databaseConn, err := dataAccess.NewSQLConnection()
	if err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}
	defer databaseConn.Destroy()

	avgTimeInTransit, err := databaseConn.ShipmentManager().GetAverageTimeInTransit(carrier)
	if err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}

	//create response
	successResponse := &struct {
		AverageTimeInTransit int    `json:"average_time_in_transit"`
		Carrier              string `json:"carrier,omitempty"`
	}{
		AverageTimeInTransit: avgTimeInTransit,
		Carrier:              carrier,
	}
	body, err := json.Marshal(successResponse)
	if err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}

	//just some info
	executionTime := time.Now().Sub(startTime)
	fmt.Printf("ExecutionTime: %s\n", executionTime)
	fmt.Printf("Average Transit Time: %v\n", avgTimeInTransit)

	return &models.APIGatewayResponse{
		StatusCode: http.StatusOK,
		Body:       string(body),
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}, nil
}

func errorResponse(code int, err error) (*models.APIGatewayResponse, error) {
	body := &struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	bodyData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return &models.APIGatewayResponse{
		StatusCode: code,
		Body:       string(bodyData),
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}, nil
}
