package integrations

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	limitedTrackingServicePath = "Prod/limited_tracking_service"
)

type WondermentAPI struct {
	baseURL string
}

func NewWondermentAPI(baseURL string) *WondermentAPI {
	return &WondermentAPI{
		baseURL: baseURL,
	}
}

func (api WondermentAPI) LimitedTrackingSerice(carrier string, trackingCode string) (*WondermentShipment, error) {
	//verify parameters
	if len(carrier) == 0 {
		return nil, errors.New("Invalid carrier")
	}
	if len(trackingCode) == 0 {
		return nil, errors.New("Invalid tracking code")
	}

	//construct url
	requestURL, err := url.Parse(api.baseURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("carrier", carrier)
	params.Add("tracking_code", trackingCode)

	requestURL.Path = limitedTrackingServicePath
	requestURL.RawQuery = params.Encode()

	//execute request
	resp, err := http.Get(requestURL.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("HTTP GET response was not OK")
	}

	//read and unmarshal body
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	shipment := &WondermentShipment{}

	err = json.Unmarshal(bodyData, shipment)
	if err != nil {
		return nil, err
	}

	return shipment, nil
}
