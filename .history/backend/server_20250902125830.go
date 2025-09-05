package main

import (
	"context"
	"database/sql"

	pb "employee_app/backend/pb"

	_ "github.com/go-sql-driver/mysql"
)

type server struct {
	pb.UnimplementedEmployeeServiceServer
	db *sql.DB
}

func NewServer(db *sql.DB) *server {
	return &server{db: db}
}

func (s *server) GetEmployees(ctx context.Context, in *pb.Empty) (*pb.EmployeeList, error) {
	rows, err := s.db.Query("SELECT id, first_name, last_name, email, position, department FROM employees")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*pb.Employee
	for rows.Next() {
		var e pb.Employee
		if err := rows.Scan(&e.Id, &e.FirstName, &e.LastName, &e.Email, &e.Position, &e.Department); err != nil {
			return nil, err
		}
		employees = append(employees, &e)
	}
	return &pb.EmployeeList{Employees: employees}, nil
}

func (s *server) CreateEmployee(ctx context.Context, in *pb.Employee) (*pb.Employee, error) {
	res, err := s.db.Exec("INSERT INTO employees (first_name, last_name, email, position, department) VALUES (?, ?, ?, ?, ?)",
		in.FirstName, in.LastName, in.Email, in.Position, in.Department)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	in.Id = int32(id)
	return in, nil
}

func (s *server) UpdateEmployee(ctx context.Context, in *pb.Employee) (*pb.Employee, error) {
	_, err := s.db.Exec("UPDATE employees SET first_name=?, last_name=?, email=?, position=?, department=? WHERE id=?",
		in.FirstName, in.LastName, in.Email, in.Position, in.Department, in.Id)
	if err != nil {
		return nil, err
	}
	return in, nil
}

func (s *server) DeleteEmployee(ctx context.Context, in *pb.EmployeeID) (*pb.Empty, error) {
	_, err := s.db.Exec("DELETE FROM employees WHERE id=?", in.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}
