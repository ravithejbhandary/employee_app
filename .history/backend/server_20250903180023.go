package main

import (
	"context"

	pb "EMPLOYEE_APP/backend/pb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedEmployeeServiceServer
	employeesCollection *mongo.Collection
}

func NewServer(collection *mongo.Collection) *server {
	return &server{employeesCollection: collection}
}

func (s *server) GetEmployees(ctx context.Context, in *pb.Empty) (*pb.EmployeeList, error) {
	var employees []*pb.Employee
	cursor, err := s.employeesCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to find employees: %v", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &employees); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to decode employees: %v", err)
	}

	return &pb.EmployeeList{Employees: employees}, nil
}

func (s *server) CreateEmployee(ctx context.Context, in *pb.Employee) (*pb.Employee, error) {
	res, err := s.employeesCollection.InsertOne(ctx, in)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create employee: %v", err)
	}

	in.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return in, nil
}

func (s *server) UpdateEmployee(ctx context.Context, in *pb.Employee) (*pb.Employee, error) {
	objID, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format")
	}

	updateDoc := bson.M{
		"$set": bson.M{
			"firstName":  in.FirstName,
			"lastName":   in.LastName,
			"email":      in.Email,
			"position":   in.Position,
			"department": in.Department,
		},
	}

	_, err = s.employeesCollection.UpdateOne(ctx, bson.M{"_id": objID}, updateDoc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update employee: %v", err)
	}

	return in, nil
}

func (s *server) DeleteEmployee(ctx context.Context, in *pb.EmployeeID) (*pb.Empty, error) {
	objID, err := primitive.ObjectIDFromHex(in.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid ID format")
	}

	_, err = s.employeesCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete employee: %v", err)
	}

	return &pb.Empty{}, nil
}
