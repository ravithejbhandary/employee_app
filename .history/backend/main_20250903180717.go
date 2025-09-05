package main

import (
	"context"
	"log"
	"os" // Import the os package
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Get MongoDB Atlas URI from environment variables
	mongoURI := os.Getenv("MONGO_ATLAS_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_ATLAS_URI environment variable not set")
	}

	// MongoDB Atlas connection
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
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

	// Start gRPC server... (rest of the code remains the same)
}
