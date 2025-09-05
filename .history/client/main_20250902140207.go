package main

import (
	"context"
	"log"
	"time"

	pb "EMPLOYEE_APP/backend/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewEmployeeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Create a new employee
	log.Println("--- Creating a new employee ---")
	newEmployee := &pb.Employee{
		FirstName:  "Jane",
		LastName:   "Doe",
		Email:      "jane.doe@example.com",
		Position:   "Software Engineer",
		Department: "Engineering",
	}
	createdEmployee, err := client.CreateEmployee(ctx, newEmployee)
	if err != nil {
		log.Fatalf("could not create employee: %v", err)
	}
	log.Printf("Created Employee: %v\n\n", createdEmployee)

	// 2. Read all employees to confirm creation
	log.Println("--- Getting all employees ---")
	employees, err := client.GetEmployees(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get employees: %v", err)
	}
	log.Printf("All Employees:\n%v\n\n", employees.Employees)

	// 3. Update the employee
	log.Println("--- Updating the employee ---")
	updatedEmployeeData := &pb.Employee{
		Id:         createdEmployee.Id, // Use the ID from the created employee
		FirstName:  "Jane",
		LastName:   "Smith",
		Email:      "jane.smith@example.com",
		Position:   "Senior Manager",
		Department: "Management",
	}
	updatedEmployee, err := client.UpdateEmployee(ctx, updatedEmployeeData)
	if err != nil {
		log.Fatalf("could not update employee: %v", err)
	}
	log.Printf("Updated Employee: %v\n\n", updatedEmployee)

	// 4. Read all employees to confirm update
	log.Println("--- Getting all employees after update ---")
	employeesAfterUpdate, err := client.GetEmployees(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get employees: %v", err)
	}
	log.Printf("All Employees After Update:\n%v\n\n", employeesAfterUpdate.Employees)

	// 5. Delete the employee
	log.Println("--- Deleting the employee ---")
	_, err = client.DeleteEmployee(ctx, &pb.EmployeeID{Id: createdEmployee.Id})
	if err != nil {
		log.Fatalf("could not delete employee: %v", err)
	}
	log.Printf("Employee with ID %d deleted successfully.\n\n", createdEmployee.Id)

	// 6. Read all employees to confirm deletion
	log.Println("--- Getting all employees after delete ---")
	employeesAfterDelete, err := client.GetEmployees(ctx, &pb.Empty{})
	if err != nil {
		log.Fatalf("could not get employees: %v", err)
	}
	log.Printf("All Employees After Delete:\n%v\n", employeesAfterDelete.Employees)
}
