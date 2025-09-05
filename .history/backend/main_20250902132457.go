package main

import (
	"database/sql"
	"log"
	"net"

	pb "EMPLOYEE_APP/backend/pb"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

func main() {
	// MySQL connection (change user:password@tcp... to your MySQL details)
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3333)/employee_db")
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterEmployeeServiceServer(s, NewServer(db))

	log.Println("gRPC server running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
