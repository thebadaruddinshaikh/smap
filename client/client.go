package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/AndrewSerra/autonomous-driving-pothole-detect/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type potholeServer struct {
	pb.UnimplementedPotholeCommunicationServiceServer
}

func (p *potholeServer) ExtendUpcomingArea(stream pb.PotholeCommunicationService_ExtendUpcomingAreaServer) error {
	for {
		request, err := stream.Recv()

		fmt.Printf("--> New Data!! -- %+v", request)

		if err == io.EOF {
			return stream.SendAndClose(&pb.ExtendUpcomingAreaResponse{
				Accepted: true,
			})
		} else if err != nil {
			fmt.Printf("could not read stream from push location: %s\n", err)
			return err
		}
	}
}

type VehicleClient struct {
	client     pb.PotholeCommunicationServiceClient
	vehicleId  string
	locationCh chan *pb.Point
	conn       *grpc.ClientConn
}

func newServer() *potholeServer {
	return &potholeServer{}
}

func (v *VehicleClient) RegisterVehicle(startLat, startLong float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := v.client.RegisterVehicle(ctx, &pb.RegisterVehicleRequest{
		StartLocation: &pb.Point{
			Latitude:  startLat,
			Longitude: startLong,
		},
		ListeningAt: "localhost:3001",
	})
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to register vehicle: %v", err)
	}

	v.vehicleId = resp.VehicleId
	log.Printf("Vehicle registered with ID: %s\n", v.vehicleId)
	return nil
}

func (v *VehicleClient) StartLocationUpdates() error {
	ctx := context.Background()

	stream, err := v.client.PushLocationUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to create location stream: %v", err)
	}

	for location := range v.locationCh {
		update := &pb.PushLocationUpdateRequest{
			VehicleId:   v.vehicleId,
			CarLocation: location,
			Potholes:    []*pb.Point{},
		}

		if err := stream.Send(update); err != nil {
			return fmt.Errorf("failed to send location update: %v", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("error receiving response: %v", err)
	}

	log.Printf("Location updates completed, accepted: %v\n", resp.Accepted)
	return nil
}

func (v *VehicleClient) UpdateLocation(lat, long float64) {
	v.locationCh <- &pb.Point{
		Latitude:  lat,
		Longitude: long,
	}
}

func NewVehicleClient(serverAddr string) (*VehicleClient, error) {
	// Create client connection using grpc.NewClient
	client, err := grpc.NewClient(serverAddr, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}...)

	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	return &VehicleClient{
		client:     pb.NewPotholeCommunicationServiceClient(client),
		locationCh: make(chan *pb.Point),
		conn:       client,
	}, nil
}

func (v *VehicleClient) Close() error {
	if v.conn != nil {
		return v.conn.Close()
	}
	return nil
}

func main() {
	// Create new vehicle client
	client, err := NewVehicleClient("localhost:3000")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Route around University of Rochester campus
	// Starting at Wilmot Building, going around campus
	coordinates := [2]float64{43.1289, -77.6298} // Wilmot Building

	// Register vehicle starting at Wilmot Building
	err = client.RegisterVehicle(coordinates[0], coordinates[1])
	if err != nil {
		log.Fatalf("Failed to register vehicle: %v", err)
	}

	// Start location updates
	go func() {
		if err := client.StartLocationUpdates(); err != nil {
			log.Printf("Location updates stopped: %v", err)
		}
	}()

	// Simulate vehicle movement around campus
	log.Println("Starting simulation around University of Rochester campus...")

	go func() {
		coordinates := [2]float64{43.1289, -77.6298}
		for {
			lat, lon := coordinates[0], coordinates[1]
			coordinates = [2]float64{lat - 0.003, lon + 0.003}
			// Update location
			client.UpdateLocation(coordinates[0], coordinates[1])
			log.Printf("Vehicle at location: Lat: %f, Long: %f\n", coordinates[0], coordinates[1])

			// Wait between updates
			time.Sleep(time.Second * 3)
		}
	}()

	lis, err := net.Listen("tcp", ":3001")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	server := newServer()
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	pb.RegisterPotholeCommunicationServiceServer(grpcServer, server)

	if err = grpcServer.Serve(lis); err != nil {
		fmt.Printf("error serving server: %s\n", err)
		os.Exit(1)
	}

	// Clean up
	// log.Println("Simulation complete")
	// close(client.locationCh)
	// time.Sleep(time.Second * 2) // Allow final updates to complete
}
