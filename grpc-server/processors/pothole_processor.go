package processors

import (
	"fmt"

	"github.com/AndrewSerra/autonomous-driving-pothole-detect/db_ops"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"
)

const ERROR_RANGE_M float32 = 3 // Estimated error range set to 3 meters

type potholeProcessor struct {
	detectionChan      chan registry.Point
	locationUpdateChan chan registry.Vehicle
	writer             db_ops.DBWriter
}

func NewPotholeProcessor(detectChan chan registry.Point, locationUpdateChan chan registry.Vehicle) *potholeProcessor {
	return &potholeProcessor{
		detectionChan:      detectChan,
		locationUpdateChan: locationUpdateChan,
		writer:             *db_ops.NewDBWriter(),
	}
}

func (p *potholeProcessor) Start() {
	for position := range p.detectionChan {
		err := p.writer.CreatePothole(position, ERROR_RANGE_M)

		if err != nil {
			fmt.Printf("could not create pothole record for point Point(%f %f): %s\n", position.Latitude, position.Longitude, err)
		}
	}
}
