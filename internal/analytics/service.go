package analytics

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsService struct {
	pool  *pgxpool.Pool
	mu    sync.RWMutex
	cache *DashboardResponse // In-memory cached result
}

func NewAnalyticsService(pool *pgxpool.Pool) *AnalyticsService {
	return &AnalyticsService{
		pool: pool,
	}
}

// Phase 4: Fan-Out / Fan-In Pattern
// Instead of fetching these 3 reports sequentially, we fire 3 goroutines (Fan-Out) 
// and aggregate the responses back into the DashboardResponse struct (Fan-In).
func (s *AnalyticsService) CalculateDashboard(ctx context.Context) (*DashboardResponse, error) {
	var wg sync.WaitGroup
	var dashboard DashboardResponse
	
	// Channels for fanning-in results
	chSpenders := make(chan []TopSpender, 1)
	chBestSellers := make(chan []BestSeller, 1)
	chPeakHours := make(chan []PeakHour, 1)
	
	errChan := make(chan error, 3)

	// Fan-Out: Task 1 (Top Spenders)
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := s.fetchTopSpenders(ctx)
		if err != nil { errChan <- err; return }
		chSpenders <- res
	}()

	// Fan-Out: Task 2 (Best Sellers)
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := s.fetchBestSellers(ctx)
		if err != nil { errChan <- err; return }
		chBestSellers <- res
	}()

	// Fan-Out: Task 3 (Peak Hours)
	wg.Add(1)
	go func() {
		defer wg.Done()
		res, err := s.fetchPeakHours(ctx)
		if err != nil { errChan <- err; return }
		chPeakHours <- res
	}()

	// Wait for all to finish
	wg.Wait()
	close(errChan)

	// Check if any error occurred
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// Fan-In: Aggregating channel responses into a single struct
	dashboard.TopSpenders = <-chSpenders
	dashboard.BestSellers = <-chBestSellers
	dashboard.PeakHours = <-chPeakHours

	return &dashboard, nil
}

// Phase 5: Background Jobs (Ticker + SQL Aggregation)
func (s *AnalyticsService) StartBackgroundJob(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	
	// Initial population of cache immediately
	s.refreshCache(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping background analytics job (Graceful shutdown).")
				ticker.Stop()
				return
			case <-ticker.C:
				log.Println("Background Job: Refreshing analytics cache...")
				s.refreshCache(ctx)
			}
		}
	}()
}

func (s *AnalyticsService) refreshCache(ctx context.Context) {
	// Call Fan-Out/Fan-In concurrent queries
	dash, err := s.CalculateDashboard(ctx)
	if err != nil {
		log.Printf("Background Job Error: failed to refresh cache: %v", err)
		return
	}
	
	// Thread-safe cache update
	s.mu.Lock()
	s.cache = dash
	s.mu.Unlock()
	log.Println("Background Job: Cache refreshed successfully.")
}

// GetCachedDashboard returns the latest aggregated data to the REST API instantly (Zero DB delay).
func (s *AnalyticsService) GetCachedDashboard() *DashboardResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache
}

// --- Raw Database Query Functions ---

func (s *AnalyticsService) fetchTopSpenders(ctx context.Context) ([]TopSpender, error) {
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
	rows, err := s.pool.Query(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	var result []TopSpender
	for rows.Next() {
		var ts TopSpender
		if err := rows.Scan(&ts.Name, &ts.TotalSpent); err != nil { return nil, err }
		result = append(result, ts)
	}
	return result, nil
}

func (s *AnalyticsService) fetchBestSellers(ctx context.Context) ([]BestSeller, error) {
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
	rows, err := s.pool.Query(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	var result []BestSeller
	for rows.Next() {
		var bs BestSeller
		if err := rows.Scan(&bs.Category, &bs.Name, &bs.TotalSold); err != nil { return nil, err }
		result = append(result, bs)
	}
	return result, nil
}

func (s *AnalyticsService) fetchPeakHours(ctx context.Context) ([]PeakHour, error) {
	query := `
		SELECT EXTRACT(HOUR FROM created_at) AS order_hour, COUNT(id) AS total_orders
		FROM orders
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY total_orders DESC
		LIMIT 5;
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	var result []PeakHour
	for rows.Next() {
		var hour float64
		var count int
		if err := rows.Scan(&hour, &count); err != nil { return nil, err }
		result = append(result, PeakHour{Hour: int(hour), TotalOrders: count})
	}
	return result, nil
}
