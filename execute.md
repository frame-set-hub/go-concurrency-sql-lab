# Execution / Progress Plan

Use this document to track project progress across phases. 

- [ ] **Phase 1: SQL Analytics (Basic to Master)**
  - [x] Initialize Database Schema (Tables: users, products, orders, order_items)
  - [x] Write Basic SQL join queries
  - [x] Write Intermediate SQL (GROUP BY & HAVING)
  - [x] Master SQL (Window functions & CTEs)
- [ ] **Phase 2: Go Routine & Concurrency (Basic)**
  - [x] Implement simple goroutine execution with WaitGroup
  - [x] Setup channels for generator-processor dummy pattern
  - [x] Fix data races using Mutex
- [ ] **Phase 3: The Integration (Intermediate)**
  - [x] Setup PostgreSQL connection in Go (`pgx`)
  - [x] Implement concurrent bulk insert (1,000,000 rows)
  - [x] Apply connection pooling settings
- [ ] **Phase 4: Master Go Concurrency Patterns**
  - [x] Implement Worker Pool Pattern (`chan` + WaitGroup)
  - [x] Implement Fan-Out/Fan-In Pattern
  - [x] Add Graceful shutdown using Context
- [ ] **Phase 5: The Grand Finale**
  - [x] Create Background Jobs (Ticker + SQL Aggregation)
  - [x] Build REST API (`gin` or native) to expose calculated analytical data
