package main

import (
	"context"
	"log"

	pb "EMPLOYEE_APP/backend/pb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MongoDB Employee model
type Employee struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	FirstName  string             `bson:"first_name"`
	LastName   string             `bson:"last_name"`
	Email      string             `bson:"email"`
	Position   string             `bson:"position"`
	Department string             `bson:"department"`
}

type server struct {
	pb.UnimplementedEmployeeServiceServer
	employeesCollection *mongo.Collection
}

func NewServer(collection *mongo.Collection) pb.EmployeeServiceServer {
	return &server{employeesCollection: collection}
}

// CreateEmployee
func (s *server) CreateEmployee(ctx context.Context, req *pb.Employee) (*pb.Employee, error) {
	log.Println("CreateEmployee RPC called")

	emp := Employee{
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Email:      req.GetEmail(),
		Position:   req.GetPosition(),
		Department: req.GetDepartment(),
	}

	res, err := s.employeesCollection.InsertOne(ctx, emp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create employee: %v", err)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Failed to get inserted ID")
	}

	req.Id = oid.Hex()
	return req, nil
}

// GetEmployees (list all)
func (s *server) GetEmployees(ctx context.Context, req *pb.Empty) (*pb.EmployeeList, error) {
	log.Println("GetEmployees RPC called")

	cursor, err := s.employeesCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve employees: %v", err)
	}
	defer cursor.Close(ctx)

	var employees []*pb.Employee
	for cursor.Next(ctx) {
		var emp Employee
		if err := cursor.Decode(&emp); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to decode employee: %v", err)
		}

		employees = append(employees, &pb.Employee{
			Id:         emp.ID.Hex(),
			FirstName:  emp.FirstName,
			LastName:   emp.LastName,
			Email:      emp.Email,
			Position:   emp.Position,
			Department: emp.Department,
		})
	}

	if err := cursor.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "Cursor error: %v", err)
	}

	return &pb.EmployeeList{Employees: employees}, nil
}

// UpdateEmployee
func (s *server) UpdateEmployee(ctx context.Context, req *pb.Employee) (*pb.Employee, error) {
	log.Println("UpdateEmployee RPC called")

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format: %v", err)
	}

	update := bson.M{"$set": Employee{
		FirstName:  req.GetFirstName(),
		LastName:   req.GetLastName(),
		Email:      req.GetEmail(),
		Position:   req.GetPosition(),
		Department: req.GetDepartment(),
	}}

	res, err := s.employeesCollection.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update employee: %v", err)
	}
	if res.MatchedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "Employee not found with ID: %s", req.GetId())
	}

	return req, nil
}

// DeleteEmployee
func (s *server) DeleteEmployee(ctx context.Context, req *pb.EmployeeID) (*pb.Empty, error) {
	log.Println("DeleteEmployee RPC called")

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format: %v", err)
	}

	res, err := s.employeesCollection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete employee: %v", err)
	}
	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "Employee not found with ID: %s", req.GetId())
	}

	return &pb.Empty{}, nil
}
