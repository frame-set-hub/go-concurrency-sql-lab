package seeder

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Seeder struct {
	pool    *pgxpool.Pool
	wg      sync.WaitGroup
	mu      sync.Mutex // For data race protection testing
	counter int
}

func NewSeeder(pool *pgxpool.Pool) *Seeder {
	gofakeit.Seed(time.Now().UnixNano())
	return &Seeder{
		pool: pool,
	}
}

// Phase 2 & 3: Worker Pool Pattern for Bulk Insert
func (s *Seeder) Run(ctx context.Context, numWorkers int, totalRecords int) {
	log.Printf("Starting data seed: %d workers, %d total order records\n", numWorkers, totalRecords)

	// We'll generate User and Product master data first (synchronously for simplicity)
	log.Println("Creating Master Data (Users, Products)...")
	users := s.seedUsers(ctx, 1000)
	products := s.seedProducts(ctx, 500)

	// Channel for distributing order generation tasks
	// Buffer size allows the generator to stay ahead of workers
	orderChan := make(chan int, 1000)

	// Start Workers (Phase 4: Worker Pool Pattern)
	for i := 1; i <= numWorkers; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i, orderChan, users, products)
	}

	// Generator (Phase 2: Generator-Processor Pattern)
	go func() {
		for i := 0; i < totalRecords; i++ {
			orderChan <- i
		}
		close(orderChan) // Signal workers to stop when channel is empty
	}()

	// Wait for all workers to finish
	s.wg.Wait()

	log.Printf("Seeding complete! Successfully processed %d orders with items.", s.counter)
}

// worker listens to the channel and inserts data using bulk copy (CopyFrom) for extreme performance.
// Or we can batch them. To make it simple per worker, we'll insert a batch of 1000 at a time.
func (s *Seeder) worker(ctx context.Context, workerID int, jobs <-chan int, users, products []string) {
	defer s.wg.Done()

	// Collecting records to insert in batch
	batchSize := 2000
	var orderRows [][]interface{}
	var orderItemRows [][]interface{}

	flush := func() {
		if len(orderRows) == 0 {
			return
		}

		// Insert Orders
		_, err := s.pool.CopyFrom(
			ctx,
			pgx.Identifier{"orders"},
			[]string{"id", "user_id", "status", "created_at"},
			pgx.CopyFromRows(orderRows),
		)
		if err != nil {
			log.Printf("Worker %d: error inserting orders: %v\n", workerID, err)
		}

		// Insert Order Items
		_, err = s.pool.CopyFrom(
			ctx,
			pgx.Identifier{"order_items"},
			[]string{"order_id", "product_id", "quantity", "unit_price"},
			pgx.CopyFromRows(orderItemRows),
		)
		if err != nil {
			log.Printf("Worker %d: error inserting order items: %v\n", workerID, err)
		}

		// Phase 2: Protecting shared map/variables using Mutex
		s.mu.Lock()
		s.counter += len(orderRows)
		s.mu.Unlock()

		orderRows = nil
		orderItemRows = nil
	}

	for range jobs {
		// Generate an Order
		orderID := gofakeit.UUID()
		userID := users[gofakeit.Number(0, len(users)-1)]
		status := randomStatus()
		createdAt := gofakeit.DateRange(time.Now().AddDate(-1, 0, 0), time.Now())

		orderRows = append(orderRows, []interface{}{orderID, userID, status, createdAt})

		// Generate 1-5 Order Items
		numItems := gofakeit.Number(1, 5)
		for i := 0; i < numItems; i++ {
			productIdx := gofakeit.Number(0, len(products)-1)
			productID := strings.Split(products[productIdx], "|")[0]
			price := gofakeit.Price(10, 500)
			qty := gofakeit.Number(1, 10)

			orderItemRows = append(orderItemRows, []interface{}{
				orderID, productID, qty, price,
			})
		}

		// Flush if batch is full
		if len(orderRows) >= batchSize {
			flush()
		}
	}

	// Final flush for remaining items
	flush()
}

// Master Data generation wrappers...
func (s *Seeder) seedUsers(ctx context.Context, count int) []string {
	var rows [][]interface{}
	var userIDs []string
	for i := 0; i < count; i++ {
		id := gofakeit.UUID()
		name := gofakeit.Name()
		createdAt := gofakeit.Date()
		userIDs = append(userIDs, id)
		rows = append(rows, []interface{}{id, name, createdAt})
	}

	_, err := s.pool.CopyFrom(ctx, pgx.Identifier{"users"}, []string{"id", "name", "created_at"}, pgx.CopyFromRows(rows))
	if err != nil {
		log.Fatalf("Error seeding users: %v", err)
	}
	return userIDs
}

func (s *Seeder) seedProducts(ctx context.Context, count int) []string {
	var rows [][]interface{}
	var products []string
	categories := []string{"Electronics", "Clothing", "Home", "Books", "Toys"}
	
	for i := 0; i < count; i++ {
		id := gofakeit.UUID()
		name := gofakeit.ProductName()
		category := categories[gofakeit.Number(0, len(categories)-1)]
		price := gofakeit.Price(10, 1000)
		products = append(products, id+"|"+category)
		rows = append(rows, []interface{}{id, name, category, price})
	}

	_, err := s.pool.CopyFrom(ctx, pgx.Identifier{"products"}, []string{"id", "name", "category", "price"}, pgx.CopyFromRows(rows))
	if err != nil {
		log.Fatalf("Error seeding products: %v", err)
	}
	return products
}

func randomStatus() string {
	statuses := []string{"PENDING", "COMPLETED", "CANCELED", "SHIPPED"}
	return statuses[gofakeit.Number(0, len(statuses)-1)]
}
