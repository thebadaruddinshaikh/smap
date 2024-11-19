package processors

import (
	"fmt"

	"github.com/AndrewSerra/autonomous-driving-pothole-detect/db_ops"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"
)

type potholeCollector struct {
	recvChan  chan registry.Vehicle
	notifChan chan Notification
	reader    db_ops.DBReader
}

func NewPotholeCollector(vehicleUpdateChan chan registry.Vehicle, notifChan chan Notification) *potholeCollector {
	return &potholeCollector{
		recvChan:  vehicleUpdateChan,
		notifChan: notifChan,
		reader:    *db_ops.NewDBReader(),
	}
}

func (n *potholeCollector) Start() {
	for vpos := range n.recvChan {
		fmt.Println(vpos)
		if registry.ShouldVehicleUpdatePotholes(vpos.Id, vpos.Location) {
			fmt.Println("update!!!")
			points, err := n.reader.FindNextBatchPotholes(vpos.Location, registry.UPDATE_RANGE)

			if err != nil {
				fmt.Printf("could not update vehicle location in registry V-%s(%f %f)\n", vpos.Id, vpos.Location.Latitude, vpos.Location.Longitude)
				continue
			}

			err = registry.UpdateLocation(vpos.Id, vpos.Location)

			if err != nil {
				fmt.Printf("could not update vehicle location in registry V-%s(%f %f)\n", vpos.Id, vpos.Location.Latitude, vpos.Location.Longitude)
				continue
			}

			for _, p := range points {
				// Send data to the notifier
				n.notifChan <- Notification{
					clientId: vpos.Id,
					point:    p,
				}
			}
		}
		fmt.Println("---- no update!!!")
	}
}
