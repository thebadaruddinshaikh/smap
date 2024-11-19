package registry

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/google/uuid"
)

const UPDATE_RANGE = 700
const TRIGGER_RANGE = 500

type Point struct {
	Latitude  float64
	Longitude float64
}

type Pothole = Point

type Vehicle struct {
	Id       string
	Location Point
}

type VehicleUpdate struct {
	Vehicle
	Discovered []Pothole
}

var locationRegistry = make(map[string]*Point)
var locationRegistryLock sync.Mutex

func SaveNewVehicle(point *Point) (string, error) {
	clientId := uuid.NewString()

	if isVehicleRegistered(clientId) {
		return "", errors.New("duplicate client id in location registry")
	}

	locationRegistryLock.Lock()
	locationRegistry[clientId] = point
	locationRegistryLock.Unlock()

	return clientId, nil
}

func RemoveVehicle(clientId string) {
	if isVehicleRegistered(clientId) {
		fmt.Printf("vehicle with clientId '%s' not registered, can't be deleted\n", clientId)
		return
	}
	locationRegistryLock.Lock()
	delete(locationRegistry, clientId)
	locationRegistryLock.Unlock()
}

func GetVehicleLocation(clientId string) *Point {
	if pos, ok := locationRegistry[clientId]; ok {
		return pos
	}
	return nil
}

func UpdateLocation(clientId string, currentLoc Point) error {
	if _, ok := locationRegistry[clientId]; !ok {
		return errors.New("clientId does not exist in locationRegistry")
	}
	locationRegistry[clientId] = &currentLoc
	return nil
}

func ShouldVehicleUpdatePotholes(clientId string, currentLoc Point) bool {
	point := GetVehicleLocation(clientId)
	return isVehicleInUpdateRange(*point, currentLoc)
}

func isVehicleRegistered(clientId string) bool {
	if clientId == "" {
		return false
	}
	_, ok := locationRegistry[clientId]
	return ok
}

func isVehicleInUpdateRange(lastUpdatePos Point, currentPos Point) bool {
	lastLat, lastLon := lastUpdatePos.Latitude, lastUpdatePos.Longitude
	currLat, currLon := currentPos.Latitude, currentPos.Longitude

	// Earth's radius in meters
	const earthRadius = 6371000.0

	// Convert latitude and longitude to radians
	lat1 := lastLat * math.Pi / 180
	lon1 := lastLon * math.Pi / 180
	lat2 := currLat * math.Pi / 180
	lon2 := currLon * math.Pi / 180

	// Differences in coordinates
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Calculate distance in meters
	distance := earthRadius * c
	fmt.Println(distance, (UPDATE_RANGE - TRIGGER_RANGE))
	return distance >= (UPDATE_RANGE - TRIGGER_RANGE)
}
