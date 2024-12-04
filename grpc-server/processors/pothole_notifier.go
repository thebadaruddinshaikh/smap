package processors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/AndrewSerra/autonomous-driving-pothole-detect/registry"
	"github.com/AndrewSerra/autonomous-driving-pothole-detect/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Notification struct {
	clientId string
	point    registry.Point
}

type clientConnection struct {
	client rpc.PotholeCommunicationServiceClient
	stream rpc.PotholeCommunicationService_ExtendUpcomingAreaClient
}

type potholeNotifier struct {
	mu           sync.RWMutex
	regMu        sync.RWMutex
	storage      map[string]*clientConnection
	notifChan    chan Notification
	registration map[string]string
}

func NewPotholeNotifier(notifChan chan Notification) *potholeNotifier {
	return &potholeNotifier{
		storage:      make(map[string]*clientConnection),
		notifChan:    notifChan,
		registration: make(map[string]string),
	}
}

func (n *potholeNotifier) Start() {
	for notif := range n.notifChan {
		n.mu.RLock()
		_, ok := n.storage[notif.clientId]
		n.mu.RUnlock()

		fmt.Printf("sending new point P(%f %f) to V-%s\n", notif.point.Latitude, notif.point.Longitude, notif.clientId)

		if !ok {
			// fmt.Printf("Cannot find connection to V-%s\n", notif.clientId)

			n.regMu.RLock()
			addr, ok := n.registration[notif.clientId]
			n.regMu.RUnlock()

			if ok {
				fmt.Printf("registering vehicle V-%s\n", notif.clientId)
				if err := n.register(notif.clientId, addr); err != nil {
					fmt.Println(err)
					continue
				}
				n.regMu.Lock()
				delete(n.registration, notif.clientId)
				n.regMu.Unlock()
			}
		}

		n.mu.RLock()
		clientConn := n.storage[notif.clientId]
		n.mu.RUnlock()

		point := &rpc.Point{
			Latitude:  notif.point.Latitude,
			Longitude: notif.point.Longitude,
		}
		fmt.Printf("sending vehicle V-%s an update\n", notif.clientId)
		err := clientConn.stream.Send(&rpc.ExtendUpcomingAreaRequest{
			Pothole: point,
		})

		if err != nil {
			log.Printf("Failed to send update to V-%s: %v\n", notif.clientId, err)
			n.handleConnectionError(notif.clientId, clientConn)
		}
	}
}

func (n *potholeNotifier) InitRegistration(clientId string, addr string) {
	n.regMu.Lock()
	n.registration[clientId] = addr
	n.regMu.Unlock()
}

func (n *potholeNotifier) register(clientId string, address string) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // Send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // Wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // Allow pings even without active streams
		}),
	}

	conn, err := grpc.NewClient(address, opts...)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to connect to vehicle %s: %v", clientId, err)
	}

	client := rpc.NewPotholeCommunicationServiceClient(conn)
	stream, err := client.ExtendUpcomingArea(context.Background())

	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create stream for client %s: %v", clientId, err)
	}

	n.mu.Lock()
	n.storage[clientId] = &clientConnection{
		client: client,
		stream: stream,
	}
	n.mu.Unlock()

	log.Printf("Successfully registered vehicle %s\n", clientId)
	return nil
}

func (n *potholeNotifier) handleConnectionError(clientId string, conn *clientConnection) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if conn != nil && conn.stream != nil {
		conn.stream.CloseSend()
	}

	delete(n.storage, clientId)
}
