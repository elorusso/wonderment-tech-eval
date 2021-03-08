package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	dataAccess "github.com/elorusso/wonderment-tech-eval/data-access"
	"github.com/elorusso/wonderment-tech-eval/integrations"
	"github.com/elorusso/wonderment-tech-eval/models"
	"golang.org/x/sync/errgroup"
)

const (
	wondermentBaseURL = "https://wrqnmf9e62.execute-api.us-east-1.amazonaws.com"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, payload *models.APIGatewayPayload) (*models.APIGatewayResponse, error) {

	startTime := time.Now()

	params := &struct {
		Carrier      string `json:"carrier"`
		TrackingCode string `json:"tracking_code"`
	}{}

	//collect parameters
	if len(payload.Body) > 0 {
		//check body
		err := json.Unmarshal([]byte(payload.Body), params)
		if err != nil {
			return errorResponse(http.StatusBadRequest, err)
		}
	} else if payload.QueryStringParameters != nil {
		//check query params
		params.Carrier = payload.QueryStringParameters["carrier"]
		params.TrackingCode = payload.QueryStringParameters["tracking_code"]
	} else {
		return errorResponse(http.StatusBadRequest, errors.New("Required parameters missing"))
	}

	if len(params.Carrier) == 0 {
		return errorResponse(http.StatusBadRequest, errors.New("Carrier paramater is required"))
	}
	if len(params.TrackingCode) == 0 {
		return errorResponse(http.StatusBadRequest, errors.New("Tracking code parameter is required"))
	}

	//fetch shipment info from Wonderment
	wondermentAPI := integrations.NewWondermentAPI(wondermentBaseURL)

	wonderShipment, err := wondermentAPI.LimitedTrackingSerice(params.Carrier, params.TrackingCode)
	if err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}

	databaseConn, err := dataAccess.NewSQLConnection()
	if err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}
	defer databaseConn.Destroy()

	shipmentManager := databaseConn.ShipmentManager()

	//save shipment, do nothing on conflict
	shipmentID, err := shipmentManager.InsertShipment(wonderShipment)
	if err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}

	firstTransitTime := time.Now()
	deliveryTime := time.Time{}

	var eg errgroup.Group
	for _, event := range wonderShipment.TrackingHistory {
		//save tracking events async, do nothing on conflict
		eventLocal := *event
		eg.Go(func() error {
			return databaseConn.TrackingEventManager().InsertTrackingEvent(eventLocal, shipmentID)
		})

		if strings.ToLower(event.Status) == "transit" && event.StatusDate.Before(firstTransitTime) {
			firstTransitTime = event.StatusDate
		} else if strings.ToLower(event.Status) == "delivered" {
			//assuming there is only one delivery event
			deliveryTime = event.StatusDate
		}
	}

	//wait for tracking events to be saved
	if err := eg.Wait(); err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}

	//calculate time in transit, if delivered
	if !deliveryTime.IsZero() && firstTransitTime != time.Now() {
		timeInTransit := deliveryTime.Sub(firstTransitTime) //nanoseconds

		err = shipmentManager.UpdateTransitTimeForShipment(shipmentID, int(timeInTransit/1000000)) //save in milliseconds
		if err != nil {
			fmt.Println(err)
			return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
		}
	}

	//just some info
	executionTime := time.Now().Sub(startTime)
	fmt.Printf("ExecutionTime: %s\n", executionTime)
	fmt.Println("ShipmentID: " + shipmentID)

	successResponse := &struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}
	body, err := json.Marshal(successResponse)
	if err != nil {
		fmt.Println(err)
		return errorResponse(http.StatusInternalServerError, errors.New("Internal Server Error"))
	}

	return &models.APIGatewayResponse{
		StatusCode: http.StatusOK,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
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
