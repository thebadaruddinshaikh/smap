package db_ops

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"

	_ "github.com/go-sql-driver/mysql"
)

type DBReader struct {
	conn *sql.DB
}

func NewDBReader() *DBReader {
	conn, err := Connect()

	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	return &DBReader{
		conn: conn,
	}
}

func (r *DBReader) FindNextBatchPotholes(point registry.Point, radius int) ([]registry.Pothole, error) {
	rows, err := r.conn.Query(GET_VEHICLE_POTHOLES_QUERY,
		point.Latitude, point.Longitude, point.Latitude, point.Longitude, radius)
	nextBatch := make([]registry.Pothole, 0)

	if err != nil {
		fmt.Printf("cannot query next pothole locations %s\n", err)
		return nil, err
	}

	for rows.Next() {
		var data struct {
			point    registry.Pothole
			distance float64
		}

		if err = rows.Scan(&data.point.Latitude, &data.point.Longitude, &data.distance); err != nil {
			fmt.Printf("cannot read pothole data from db: %s\n", err)
			return nil, err
		}

		nextBatch = append(nextBatch, data.point)
	}

	return nextBatch, nil
}
