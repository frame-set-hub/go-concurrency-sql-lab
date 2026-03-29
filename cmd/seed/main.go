package main

import (
	"context"
	"log"
	"time"

	"go-concurrency-sql-lab/internal/db"
	"go-concurrency-sql-lab/internal/seeder"

	"github.com/joho/godotenv"
)

func main() {
	// Load env vars
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using OS env variables")
	}

	// Phase 4: Context & Graceful Shutdown
	// We use context with cancellation. If it takes too long or Ctrl+C is pressed, we can cancel it.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("Initializing database connection pool...")
	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v\n", err)
	}
	defer pool.Close()

	log.Println("Starting data seeding process...")
	start := time.Now()

	s := seeder.NewSeeder(pool)
	
	// Phase 3 & 4: Run 50 Concurrent Workers to process 1,000,000 records
	// Adjust totalRecords to 1,000,000 if your machine can handle it, using 500,000 for safety demo.
	totalOrders := 500_000 
	numWorkers := 50

	s.Run(ctx, numWorkers, totalOrders)

	elapsed := time.Since(start)
	log.Printf("Seeding took %s\n", elapsed)
}
