package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	dataAccess "github.com/elorusso/wonderment-tech-eval/data-access"
	"github.com/elorusso/wonderment-tech-eval/integrations"
	"github.com/elorusso/wonderment-tech-eval/models"
)

const (
	wondermentBaseURL = "https://wrqnmf9e62.execute-api.us-east-1.amazonaws.com/Prod"
)

func main() {
	HandleRequest(context.Background(), &models.APIGatewayPayload{
		QueryStringParameters: map[string]string{
			"carrier":       "usps",
			"tracking_code": "9400111899223817576195",
		},
	})
	// lambda.Start(HandleRequest)
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
			return &models.APIGatewayResponse{
				StatusCode: http.StatusBadRequest,
			}, err
		}
	} else if payload.QueryStringParameters != nil {
		//check query params
		carrier, ok := payload.QueryStringParameters["carrier"]
		if !ok || len(carrier) == 0 {
			return &models.APIGatewayResponse{
				StatusCode: http.StatusBadRequest,
			}, errors.New("Carrier paramater is required")
		}
		trackingCode, ok := payload.QueryStringParameters["tracking_code"]
		if !ok || len(trackingCode) == 0 {
			return &models.APIGatewayResponse{
				StatusCode: http.StatusBadRequest,
			}, errors.New("Tracking code parameter is required")
		}

		params.Carrier = carrier
		params.TrackingCode = trackingCode
	} else {
		return &models.APIGatewayResponse{
			StatusCode: http.StatusBadRequest,
		}, errors.New("Required parameters missing")
	}

	//fetch shipment info from Wonderment
	wondermentAPI := integrations.NewWondermentAPI(wondermentBaseURL)

	wonderShipment, err := wondermentAPI.LimitedTrackingSerice(params.Carrier, params.TrackingCode)
	if err != nil {
		return &models.APIGatewayResponse{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	databaseConn, err := dataAccess.NewSQLConnection()
	if err != nil {
		return nil, err
	}

	shipmentManager := databaseConn.ShipmentManager()
	trackingManager := databaseConn.TrackingEventManager()

	//save shipment, do nothing on conflict
	shipmentID, err := shipmentManager.InsertShipment(wonderShipment)
	if err != nil {
		return nil, err
	}

	firstTransitTime := time.Now()
	deliveryTime := time.Time{}
	for _, event := range wonderShipment.TrackingHistory {
		//save tracking events, do nothing on conflict
		trackingManager.InsertTrackingEvent(event, shipmentID)

		if strings.ToLower(event.Status) == "transit" && event.StatusDate.Before(firstTransitTime) {
			firstTransitTime = event.StatusDate
		} else if strings.ToLower(event.Status) == "delivered" {
			//assuming there is only one delivery event
			deliveryTime = event.StatusDate
		}
	}

	//calculate time in transit, if delivered
	if !deliveryTime.IsZero() && firstTransitTime != time.Now() {
		timeInTransit := deliveryTime.Sub(firstTransitTime) //nanoseconds

		err = shipmentManager.UpdateTransitTimeForShipment(shipmentID, int(timeInTransit/1000000)) //save in milliseconds
		if err != nil {
			return nil, err
		}
	}

	//TODO: update error returns

	executionTime := time.Now().Sub(startTime)
	fmt.Printf("ExecutionTime: %s\n", executionTime)
	fmt.Println("ShipmentID: " + shipmentID)
	return nil, nil
}
