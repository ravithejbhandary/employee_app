package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "EMPLOYEE_APP/backend/pb"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewServer is a constructor function for our gRPC server
func NewServer(collection *mongo.Collection) *server {
	return &server{employeesCollection: collection}
}

func main() {
	// Context for database connection with a 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use the Docker Compose service name "mongo-db" to connect to the local MongoDB container
	mongoURI := "mongodb://mongo-db:27017"

	// Connect to the local MongoDB instance
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// Ping the database to confirm the connection is successful
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	fmt.Println("Successfully connected to local MongoDB!")

	// Get a handle to the database and collection
	employeesCollection := client.Database("employee_db").Collection("employees")
	if employeesCollection == nil {
		log.Fatal("Failed to get employees collection.")
	}

	// Create a listener on the gRPC port
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	s := grpc.NewServer()

	// Register the service with our custom server implementation
	pb.RegisterEmployeeServiceServer(s, NewServer(employeesCollection))

	// Register reflection service on gRPC server for debugging with grpcurl
	reflection.Register(s)

	log.Printf("gRPC server listening on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
