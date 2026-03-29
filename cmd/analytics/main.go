package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"go-concurrency-sql-lab/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("🚀 Running SQL Analytics Engine...")
	fmt.Println("---------------------------------------------------")

	getTopSpenders(ctx, pool)
	getBestSellersByCategory(ctx, pool)
	getPeakHours(ctx, pool)
}

func getTopSpenders(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n📊 1️⃣ Top Spenders (> $10,000)")
	query := `
		SELECT u.name, SUM(oi.quantity * oi.unit_price) AS total_spent
		FROM users u
		JOIN orders o ON u.id = o.user_id
		JOIN order_items oi ON o.id = oi.order_id
		GROUP BY u.id, u.name
		HAVING SUM(oi.quantity * oi.unit_price) > 10000
		ORDER BY total_spent DESC
		LIMIT 5;
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Name\t Total Spent ($)\t")
	for rows.Next() {
		var name string
		var totalSpent float64
		rows.Scan(&name, &totalSpent)
		fmt.Fprintf(w, "%s\t %.2f\t\n", name, totalSpent)
	}
	w.Flush()
}

func getBestSellersByCategory(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n📦 2️⃣ Best Sellers by Category")
	query := `
		WITH CategorySales AS (
			SELECT p.category, p.name, SUM(oi.quantity) as total_sold
			FROM products p
			JOIN order_items oi ON p.id = oi.product_id
			GROUP BY p.category, p.name
		),
		RankedSales AS (
			SELECT category, name, total_sold,
				   RANK() OVER(PARTITION BY category ORDER BY total_sold DESC) as rank
			FROM CategorySales
		)
		SELECT category, name, total_sold
		FROM RankedSales
		WHERE rank = 1
		ORDER BY category;
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Category\t Product Name\t Total Sold (Units)\t")
	for rows.Next() {
		var cat, name string
		var sold int
		rows.Scan(&cat, &name, &sold)
		fmt.Fprintf(w, "%s\t %s\t %d\t\n", cat, name, sold)
	}
	w.Flush()
}

func getPeakHours(ctx context.Context, pool *pgxpool.Pool) {
	fmt.Println("\n🕒 3️⃣ Peak Ordering Hours")
	query := `
		SELECT EXTRACT(HOUR FROM created_at) AS order_hour, COUNT(id) AS total_orders
		FROM orders
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY total_orders DESC
		LIMIT 5;
	`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Hour of Day\t Total Orders Placed\t")
	for rows.Next() {
		var hour float64
		var count int
		rows.Scan(&hour, &count)
		fmt.Fprintf(w, "%02d:00 - %02d:59\t %d\t\n", int(hour), int(hour), count)
	}
	w.Flush()
}
