package analytics

// TopSpender represents a user who spent over the threshold
type TopSpender struct {
	Name       string  `json:"name"`
	TotalSpent float64 `json:"total_spent"`
}

// BestSeller represents a top product in its category
type BestSeller struct {
	Category  string `json:"category"`
	Name      string `json:"name"`
	TotalSold int    `json:"total_sold"`
}

// PeakHour represents an hour with high order volume
type PeakHour struct {
	Hour        int `json:"hour"`
	TotalOrders int `json:"total_orders"`
}

// DashboardResponse represents the aggregated JSON response (Fan-in Target)
type DashboardResponse struct {
	TopSpenders []TopSpender `json:"top_spenders"`
	BestSellers []BestSeller `json:"best_sellers"`
	PeakHours   []PeakHour   `json:"peak_hours"`
}
