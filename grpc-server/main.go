package main

import (
	"fmt"
	"log"
	"net"

	"github.com/AndrewSerra/autonomous-driving-pothole-detect/db_ops"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/processors"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/rpc"
	"google.golang.org/grpc"
)

func main() {
	port := 3000
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err = db_ops.InitiateDatabase(); err != nil {
		log.Fatalf("cannot initiate database %s\n", err)
	}

	var opts []grpc.ServerOption
	var vehicleUpdateChan chan registry.VehicleUpdate = make(chan registry.VehicleUpdate)
	var potholeDetectChan chan registry.Point = make(chan registry.Point)
	var vehicleLocationChan chan registry.Vehicle = make(chan registry.Vehicle)
	var notifChan chan processors.Notification = make(chan processors.Notification)

	locationProc := processors.NewLocationProcessor(potholeDetectChan, vehicleUpdateChan, vehicleLocationChan)
	detectedPotholeProc := processors.NewPotholeProcessor(potholeDetectChan, vehicleLocationChan)
	potholeNotifier := processors.NewPotholeNotifier(notifChan)
	pointCollector := processors.NewPotholeCollector(vehicleLocationChan, notifChan)

	potholeServer := NewPotholeServer(potholeDetectChan, vehicleUpdateChan, potholeNotifier.InitRegistration)

	grpcServer := grpc.NewServer(opts...)
	rpc.RegisterPotholeCommunicationServiceServer(grpcServer, potholeServer)

	go locationProc.Start()
	go detectedPotholeProc.Start()
	go potholeNotifier.Start()
	go pointCollector.Start()

	if err = grpcServer.Serve(lis); err != nil {
		fmt.Printf("error serving server: %s\n", err)
	}
}
