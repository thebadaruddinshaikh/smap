# Autonomous Driving Pothole Crowdsourcing

This code is a backend service that creates a bidi-streaming system. The goal is for vehicles to push their location upstream and track their position to determine whether sending locations of known potholes as they drive is needed. The protocol used is gRPC. The goal is for either a smartphone or a vehicle to run a system and notify this server to store potholes and receive real-time updates. 

## System Components

The system consists of several components that are independent of each other. Each request for a location update is passed on to a point collector which is responsible of reading the database and checking if there is a need to send new points to the vehicle. Along with a vehicle position update, any detected new pothole locations are sent along with it. If there are new potholes sent, these are sent to a pothole processor. The pothole processor checks if the points were possibly seen before and adds them to the database. Finally, if a vehicle needs pothole updates, they are passed on to the position notifier which sends a vehicle the position of a known pothole.

<img width="1334" alt="system" src="https://github.com/user-attachments/assets/0ca12a95-1bff-46ab-abd2-8eba1b63d461">

### Location Registry

The location registry is where vehicle ids are stored and their latest known update position. Once a vehicle sends a request to register to the system, the notifier will send a request to initiate a stream. The latest known update position is marked in the location registry and will assess if the defined hardcoded update range minus the trigger range is reached.

The "update range" is the circular range in meters to find known potholes with respect to the vehicles current position. The "trigger range" is defined as a measure of how close to the end of the learned circular region should the vehicle start receiving the next circular region.

## How to run

There are two files to run. These two files are `main.go` and `server.go`. The command below will run the server communicating with the database. 

```
go run main.go server.go
```

Running this command will also make the backend create the table, index, and trigger in the database. The file can be found in `db_ops/structure.sql`.


