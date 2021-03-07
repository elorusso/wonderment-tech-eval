package integrations

import (
	"time"
)

type WondermentShipment struct {
	TrackingNumber  string `json:"tracking_number"`
	Carrier         string
	ServiceLevel    *ServiceLevel
	AddressFrom     *Address `json:"address_from"`
	AddressTo       *Address `json:"address_to"`
	ETA             time.Time
	OriginalETA     time.Time `json:"original_eta"`
	Test            bool
	TrackingStatus  *TrackingEvent   `json:"tracking_status"`
	TrackingHistory []*TrackingEvent `json:"tracking_history"`
	Messages        []string
}

type TrackingEvent struct {
	StatusDate    time.Time `json:"status_date"`
	StatusDetails *string   `json:"status_details"`
	Location      *Address
	SubStatus     *SubStatus
	Created       time.Time `json:"object_created"`
	Updated       time.Time `json:"object_updated"`
	EventID       string    `json:"object_id"`
	Status        string
}

type SubStatus struct {
	Code           *string
	Text           *string
	ActionRequired bool
}

type ServiceLevel struct {
	Name  *string
	Token *string
}

type Address struct {
	City    *string
	State   *string
	Zip     *string
	Country *string
}
