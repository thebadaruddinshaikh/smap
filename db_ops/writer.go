package db_ops

import (
	"database/sql"
	"errors"
	"log"

	"github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"
	_ "github.com/go-sql-driver/mysql"
)

type DBWriter struct {
	conn *sql.DB
}

func NewDBWriter() *DBWriter {

	conn, err := Connect()

	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	return &DBWriter{
		conn: conn,
	}
}

func (w *DBWriter) CreatePothole(point registry.Point, _range float32) error {
	res, err := w.conn.Exec(INSERT_POTHOLE_QUERY, point.Latitude, point.Longitude, _range)

	if err != nil {
		return err
	}

	affectedCnt, err := res.RowsAffected()

	if err != nil {
		return err
	} else if affectedCnt == 0 {
		return errors.New("possible duplicate point")
	}

	return nil
}
