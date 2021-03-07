package dataAccess

import (
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/elorusso/wonderment-tech-eval/integrations"
)

const (
	trackingEventTableName = "tracking_events"
)

type TrackingEventManager struct {
	dbHelper *sql.DB
}

func (man TrackingEventManager) InsertTrackingEvent(event *integrations.TrackingEvent, shipmentID string) error {
	if event == nil {
		return errors.New("nil tracking event")
	}
	if len(shipmentID) == 0 {
		return errors.New("invalid shipment ID")
	}

	if event.Location == nil {
		event.Location = &integrations.Address{}
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sql, args, err := psql.Insert(trackingEventTableName).
		Columns(
			"event_id",
			"status_date",
			"status_details",
			"location_city",
			"location_state",
			"location_zip",
			"location_country",
			"substatus_code",
			"substatus_text",
			"substatus_action_required",
			"status",
			"shipment_id").
		Values(
			event.EventID,
			event.StatusDate,
			event.StatusDetails,
			event.Location.City,
			event.Location.State,
			event.Location.Zip,
			event.Location.Country,
			event.SubStatus.Code,
			event.SubStatus.Text,
			event.SubStatus.ActionRequired,
			event.Status,
			shipmentID,
		).
		Suffix("ON CONFLICT (event_id) DO NOTHING").
		ToSql()
	if err != nil {
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
