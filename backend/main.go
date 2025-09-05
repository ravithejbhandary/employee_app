package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	pb "EMPLOYEE_APP/backend/pb"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// MongoDB connection
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo-db:27017"))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	// Get a handle for the employees collection
	db := client.Database("employee_db")
	employeesCollection := db.Collection("employees")

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterEmployeeServiceServer(grpcServer, NewServer(employeesCollection))

	go func() {
		log.Println("gRPC server running on port 50051...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start gRPC-Gateway server (REST proxy)
	mux := runtime.NewServeMux()
	err = pb.RegisterEmployeeServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		":50051", // ✅ Use container-local port, not localhost
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		log.Fatalf("Failed to register gRPC-Gateway: %v", err)
	}

	log.Println("HTTP gateway running on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
