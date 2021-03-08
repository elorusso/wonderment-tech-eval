package dataAccess

import (
	"database/sql"
	"fmt"
)

const (
	host     = "wanderment-eval.czdy1hrxxpzb.us-east-1.rds.amazonaws.com"
	port     = 5432
	user     = "postgres"
	password = "masterpassword"
	dbname   = "wonderment"
)

type SQLConnection struct {
	dbHelper *sql.DB
}

func NewSQLConnection() (*SQLConnection, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	return &SQLConnection{
		dbHelper: db,
	}, nil
}

func (conn *SQLConnection) Destroy() {
	conn.dbHelper.Close()
}

func (conn SQLConnection) ShipmentManager() *ShipmentsManager {
	return &ShipmentsManager{
		dbHelper: conn.dbHelper,
	}
}

func (conn SQLConnection) TrackingEventManager() *TrackingEventManager {
	return &TrackingEventManager{
		dbHelper: conn.dbHelper,
	}
}
