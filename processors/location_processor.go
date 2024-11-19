package processors

import (
	"sync"

	"github.com/AndrewSerra/autonomous-driving-pothole-detect/db_ops"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"
)

type locationProcessor struct {
	detectionChan      chan registry.Point
	vehicleUpdateChan  chan registry.VehicleUpdate
	locationUpdateChan chan registry.Vehicle
	reader             db_ops.DBReader
}

func NewLocationProcessor(detectChan chan registry.Point, vehicleUpdateChan chan registry.VehicleUpdate, locUpdateChan chan registry.Vehicle) *locationProcessor {
	return &locationProcessor{
		detectionChan:      detectChan,
		vehicleUpdateChan:  vehicleUpdateChan,
		locationUpdateChan: locUpdateChan,
		reader:             *db_ops.NewDBReader(),
	}
}

func (p *locationProcessor) Start() {
	var wg sync.WaitGroup
	// Receive positions from the incoming request
	for data := range p.vehicleUpdateChan {
		wg.Add(1)
		go func() {
			for _, d := range data.Discovered {
				p.detectionChan <- d
			}
			wg.Done()
		}()

		p.locationUpdateChan <- data.Vehicle // Updates the location
		wg.Wait()
	}
}
