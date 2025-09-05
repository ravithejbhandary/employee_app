package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	pb "EMPLOYEE_APP/backend/pb"

	_ "github.com/go-sql-driver/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	// MySQL connection (adjust if needed for Docker)
	db, err := sql.Open("mysql", "root:root@tcp(mysql-db:3306)/employee_db")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterEmployeeServiceServer(grpcServer, NewServer(db))

	go func() {
		log.Println("gRPC server running on port 50051...")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start gRPC-Gateway server (REST proxy)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	err = pb.RegisterEmployeeServiceHandlerFromEndpoint(
		ctx, mux, "localhost:50051",
		[]grpc.DialOption{grpc.WithInsecure()},
	)
	if err != nil {
		log.Fatalf("Failed to register gRPC-Gateway: %v", err)
	}

	log.Println("HTTP gateway running on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to serve HTTP: %v", err)
	}
}
