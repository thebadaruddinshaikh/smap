package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	r "github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/rpc"
)

type PotholeServer struct {
	rpc.UnimplementedPotholeCommunicationServiceServer
	detectionChan     chan r.Point
	vehicleUpdateChan chan r.VehicleUpdate
	register          func(string, string)
}

func NewPotholeServer(detectChan chan r.Point, updateChan chan r.VehicleUpdate, registerFunc func(string, string)) *PotholeServer {
	return &PotholeServer{
		detectionChan:     detectChan,
		vehicleUpdateChan: updateChan,
		register:          registerFunc,
	}
}

func (s *PotholeServer) RegisterVehicle(ctx context.Context, in *rpc.RegisterVehicleRequest) (*rpc.RegisterVehicleResponse, error) {

	point := r.Point{
		Latitude:  in.StartLocation.GetLatitude(),
		Longitude: in.StartLocation.GetLongitude(),
	}

	vehicleId, err := r.SaveNewVehicle(&point)

	if err != nil {
		return nil, errors.New("cannot save new vehicle")
	}

	s.register(vehicleId, in.GetListeningAt())

	return &rpc.RegisterVehicleResponse{
		VehicleId: vehicleId,
	}, nil
}

func (s *PotholeServer) PushLocationUpdate(stream rpc.PotholeCommunicationService_PushLocationUpdateServer) error {
	for {
		request, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&rpc.PushLocationUpdateResponse{
				Accepted: true,
			})
		} else if err != nil {
			fmt.Printf("could not read stream from push location: %s\n", err)
			return err
		}

		// Validate the request
		if request == nil {
			return errors.New("received nil request")
		}

		if request.CarLocation == nil {
			return errors.New("received nil car location")
		}

		point := r.VehicleUpdate{
			Vehicle: r.Vehicle{
				Id: request.GetVehicleId(),
				Location: r.Point{
					Latitude:  request.CarLocation.GetLatitude(),
					Longitude: request.CarLocation.GetLongitude(),
				},
			},
			Discovered: func() []r.Point {
				locations := request.GetPotholes()
				potholes := make([]r.Point, len(locations))
				for i, l := range locations {
					potholes[i] = r.Point{
						Latitude:  l.GetLatitude(),
						Longitude: l.GetLongitude(),
					}
				}
				return potholes
			}(),
		}

		s.vehicleUpdateChan <- point
	}
}

func (s *PotholeServer) ExtendUpcomingArea(stream rpc.PotholeCommunicationService_ExtendUpcomingAreaServer) error {
	return errors.New("unexpected call of ExtendUpcomingArea from client")
}
