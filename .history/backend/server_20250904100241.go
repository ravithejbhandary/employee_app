package main

import (
	"context"
	"errors"
	"log"

	pb "employee_app/backend/pb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// server is the gRPC server implementation. It holds a reference to the MongoDB collection.
type server struct {
	pb.UnimplementedEmployeeServiceServer
	employeesCollection *mongo.Collection
}

// NewServer creates a new gRPC server instance with a MongoDB collection.
func NewServer(collection *mongo.Collection) pb.EmployeeServiceServer {
	return &server{employeesCollection: collection}
}

// CreateEmployee inserts a new employee into the MongoDB collection.
func (s *server) CreateEmployee(ctx context.Context, req *pb.CreateEmployeeRequest) (*pb.CreateEmployeeResponse, error) {
	log.Println("CreateEmployee RPC called")
	employee := req.GetEmployee()

	// Convert the proto struct to a BSON document for MongoDB
	res, err := s.employeesCollection.InsertOne(ctx, bson.M{
		"first_name": employee.GetFirstName(),
		"last_name":  employee.GetLastName(),
		"email":      employee.GetEmail(),
		"position":   employee.GetPosition(),
		"department": employee.GetDepartment(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create employee: %v", err)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Failed to get inserted ID")
	}

	employee.Id = oid.Hex()
	return &pb.CreateEmployeeResponse{Employee: employee}, nil
}

// GetEmployee retrieves a single employee by their ID.
func (s *server) GetEmployee(ctx context.Context, req *pb.GetEmployeeRequest) (*pb.GetEmployeeResponse, error) {
	log.Println("GetEmployee RPC called")
	id := req.GetId()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format: %v", err)
	}

	var employee pb.Employee
	err = s.employeesCollection.FindOne(ctx, bson.M{"_id": oid}).Decode(&employee)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, status.Errorf(codes.NotFound, "Employee not found with ID: %s", id)
		}
		return nil, status.Errorf(codes.Internal, "Failed to find employee: %v", err)
	}

	// Set the ID field correctly after decoding from MongoDB's _id
	employee.Id = oid.Hex()

	return &pb.GetEmployeeResponse{Employee: &employee}, nil
}

// UpdateEmployee updates an existing employee.
func (s *server) UpdateEmployee(ctx context.Context, req *pb.UpdateEmployeeRequest) (*pb.UpdateEmployeeResponse, error) {
	log.Println("UpdateEmployee RPC called")
	employee := req.GetEmployee()

	oid, err := primitive.ObjectIDFromHex(employee.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format: %v", err)
	}

	filter := bson.M{"_id": oid}
	update := bson.M{"$set": bson.M{
		"first_name": employee.GetFirstName(),
		"last_name":  employee.GetLastName(),
		"email":      employee.GetEmail(),
		"position":   employee.GetPosition(),
		"department": employee.GetDepartment(),
	}}

	res, err := s.employeesCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update employee: %v", err)
	}
	if res.MatchedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "Employee not found with ID: %s", employee.GetId())
	}

	return &pb.UpdateEmployeeResponse{Employee: employee}, nil
}

// DeleteEmployee deletes an employee by their ID.
func (s *server) DeleteEmployee(ctx context.Context, req *pb.DeleteEmployeeRequest) (*pb.DeleteEmployeeResponse, error) {
	log.Println("DeleteEmployee RPC called")
	id := req.GetId()
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format: %v", err)
	}

	res, err := s.employeesCollection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete employee: %v", err)
	}
	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "Employee not found with ID: %s", id)
	}

	return &pb.DeleteEmployeeResponse{Success: true}, nil
}

// ListEmployees retrieves all employees from the collection.
func (s *server) ListEmployees(ctx context.Context, req *pb.ListEmployeesRequest) (*pb.ListEmployeesResponse, error) {
	log.Println("ListEmployees RPC called")
	cursor, err := s.employeesCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve employees: %v", err)
	}
	defer cursor.Close(ctx)

	var employees []*pb.Employee
	for cursor.Next(ctx) {
		var employee pb.Employee
		if err := cursor.Decode(&employee); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to decode employee: %v", err)
		}

		// Decode the _id to the correct ID format for the proto message
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to decode raw document for ID: %v", err)
		}
		if oid, ok := doc["_id"].(primitive.ObjectID); ok {
			employee.Id = oid.Hex()
		}

		employees = append(employees, &employee)
	}

	if err := cursor.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "Cursor error: %v", err)
	}

	return &pb.ListEmployeesResponse{Employees: employees}, nil
}
