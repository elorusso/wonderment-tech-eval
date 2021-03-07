package dataAccess

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/elorusso/wonderment-tech-eval/integrations"
	_ "github.com/lib/pq"
)

const (
	shipmentsTableName = "shipments"
)

type ShipmentsManager struct {
	dbHelper *sql.DB
}

//InsertShipment creates a new shipment in the database and returns the shipment ID. If the shipment already exisits, the existing shipment ID is returned.
func (man ShipmentsManager) InsertShipment(shipment *integrations.WondermentShipment) (string, error) {
	if shipment == nil {
		return "", errors.New("nil shipment")
	}

	var etaString *string
	if !shipment.ETA.IsZero() {
		temp := shipment.ETA.Format(time.RFC3339)
		etaString = &temp
	}

	var originalETAString *string
	if !shipment.OriginalETA.IsZero() {
		temp := shipment.OriginalETA.Format(time.RFC3339)
		originalETAString = &temp
	}

	if shipment.AddressFrom == nil {
		shipment.AddressFrom = &integrations.Address{}
	}
	if shipment.AddressTo == nil {
		shipment.AddressTo = &integrations.Address{}
	}
	if shipment.ServiceLevel == nil {
		shipment.ServiceLevel = &integrations.ServiceLevel{}
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	//build sql and execute
	sql, args, err := psql.Insert(shipmentsTableName).Columns(
		"tracking_number",
		"carrier",
		"service_level_name",
		"service_level_token",
		"address_from_city",
		"address_from_state",
		"address_from_zip",
		"address_from_country",
		"address_to_city",
		"address_to_state",
		"address_to_zip",
		"address_to_country",
		"test",
		"eta",
		"original_eta").
		Values(
			shipment.TrackingNumber,
			shipment.Carrier,
			shipment.ServiceLevel.Name,
			shipment.ServiceLevel.Token,
			shipment.AddressFrom.City,
			shipment.AddressFrom.State,
			shipment.AddressFrom.Zip,
			shipment.AddressFrom.Country,
			shipment.AddressTo.City,
			shipment.AddressTo.State,
			shipment.AddressTo.Zip,
			shipment.AddressTo.Country,
			shipment.Test,
			etaString,
			originalETAString).
		Suffix(
			"ON CONFLICT (carrier, tracking_number) DO UPDATE SET carrier=EXCLUDED.carrier RETURNING shipment_id"). //make sure we get a shipment ID back even on conflict
		ToSql()
	if err != nil {
		return "", err
	}

	fmt.Println(sql, args)

	rows, err := man.dbHelper.Query(sql, args...)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var shipmentID string

	if rows.Next() {
		err = rows.Scan(&shipmentID)
		if err != nil {
			return "", err
		}
	} else if rows.Err() != nil {
		return "", err
	} else {
		return "", errors.New("Failed to get expected shipment ID")
	}
	return shipmentID, nil
}

func (man ShipmentsManager) UpdateTransitTimeForShipment(shipmentID string, transitTime int) error {
	if len(shipmentID) == 0 {
		return errors.New("Invalid shipment ID")
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := psql.Update(shipmentsTableName).Set("time_in_transit", transitTime).Where(sq.Eq{"shipment_id": shipmentID}).ToSql()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(sql, args)

	_, err = man.dbHelper.Query(sql, args...)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
