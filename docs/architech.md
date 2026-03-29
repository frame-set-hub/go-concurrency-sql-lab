# Architecture & System Diagrams

เรื่อง Concurrency และ Database เป็นหัวข้อที่ซับซ้อน การมองภาพรวมให้ออกจะช่วยให้เข้าใจง่ายขึ้นมาก หน้านี้รวบรวม Diagram ต่างๆ ที่อธิบายระบบต่างๆ ในโปรเจกต์นี้ไว้ทั้งหมดครับ โดยใช้ **Mermaid** ในการเรนเดอร์ภาพ

---

## 🏗️ 1. Go Concurrency Architecture (Worker Pool Pattern)

ไดอะแกรมนี้อธิบายการทำงานของ **Phase 2 & Phase 3** ภาพรวมของโกรูทีน (Goroutines) และการรับส่งข้อมูลผ่าน `Channel` จะเป็นรูปแบบ **Generator-Processor (Worker Pool)**:

- **Generator (Main)**: หน้าที่คอยโยน "เลขงาน" (Job ID) ลงไปในกล่อง (Channel) 
- **Channel**: ท่อส่งข้อมูลที่เป็น Buffer ทำให้ Main รันต่อไปได้โดยไม่ต้องรอ Worker ทุกตัวพร้อม
- **Workers (Goroutines 1 to 50)**: โกรูทีนลูกข่าย 50 ตัววิ่งไปหยิบงานออกจากกล่อง มาสุ่มสร้างข้อมูล `orders` และ `order_items`
- **Mutex**: เมื่อ Worker ทำงานเสร็จ จะไปอัปเดต Counter รวม โดยต้องทำการ Lock `Mutex` เพื่อไม่ให้เกิด Data Race (แย่งกันเขียนข้อมูล)
- **Batching & Bulk Insert**: ยัดข้อมูลทีละ 2,000 ต่อ Worker ช่วยลดเวลาในการต่อท่อเข้า Database อย่างมหาศาล

```mermaid
flowchart TD
    Main[Main Thread & Generator] -->|1. Generate Jobs 0 to 500k| JobChan[(Order Job Channel)]
    
    subgraph WaitGroup - 50 Workers
        W1((Worker 1))
        W2((Worker 2))
        Wdot((... Worker 50))
    end
    
    JobChan -->|2. Pull Job| W1
    JobChan -->|2. Pull Job| W2
    JobChan -->|2. Pull Job| Wdot
    
    W1 -->|3. Generate Fake Data| Batch1[Local Array Batch]
    W2 -->|3. Generate Fake Data| Batch2[Local Array Batch]
    
    Batch1 -->|4. CopyFrom Insert| DB[(PostgreSQL)]
    Batch2 -->|4. CopyFrom Insert| DB
    
    W1 -.->|5. Lock Mutex| Counter{Global Counter}
    W2 -.->|5. Lock Mutex| Counter
```

---

## 🗄️ 2. Database Schema (ER Diagram)

ภาพรวมของตารางทั้งหมดภายใน PostgreSQL ที่ถูกสร้างจาก `Phase 1` เป็นโครงสร้างคลาสสิกของระบบ **E-Commerce** เราจะใช้ตารางเหล่านี้ให้เป็นประโยชน์หนักๆ ในการทำ Query `GROUP BY` ของ Phase ถัดๆ ไป

```mermaid
erDiagram
    USERS ||--o{ ORDERS : "places"
    PRODUCTS ||--o{ ORDER_ITEMS : "included in"
    ORDERS ||--o{ ORDER_ITEMS : "contains"

    USERS {
        uuid id PK
        varchar name
        timestamp created_at
    }
    
    PRODUCTS {
        uuid id PK
        varchar name
        varchar category
        decimal price
    }
    
    ORDERS {
        uuid id PK
        uuid user_id FK
        varchar status "e.g., PENDING, COMPLETED"
        timestamp created_at
    }
    
    ORDER_ITEMS {
        uuid id PK
        uuid order_id FK
        uuid product_id FK
        int quantity
        decimal unit_price
    }
```

---

## ⏱️ 3. Execution Flow (Sequence Diagram)

การไล่ลำดับการทำงาน (Sequence) ตั้งแต่เราสั่ง `go run cmd/seed/main.go` จนโปรแกรมทำงานเสร็จ จะเห็นได้ว่าจังหวะที่ Worker 50 ตัวทำงาน มันไม่ได้ทำแบบเส้นตรง (Synchronous) แต่ทำพร้อมกันทั้งหมด (Asynchronous/Concurrent) 

```mermaid
sequenceDiagram
    participant Main
    participant Context
    participant DBPool
    participant Generator
    participant Worker (x50)
    participant Postgres

    Main->>Context: Create with Timeout (5 mins)
    Main->>DBPool: Setup Connection Pool (Max=100)
    
    Main->>Postgres: Bulk Insert Master Data (Users/Products)
    Postgres-->>Main: Success
    
    Main->>Generator: Start Goroutine
    
    par For 1 to 50
        Main->>Worker (x50): wg.Add(1) & Start Goroutine
    end

    loop Total 500,000 times
        Generator->>Worker (x50): Send Job (Channel)
    end
    Generator->>Worker (x50): close(channel)
    
    loop Every 2000 items (per worker)
        Worker (x50)->>Postgres: pgx.CopyFrom (Orders + Items)
        Postgres-->>Worker (x50): Inserted
        Worker (x50)->>Worker (x50): mu.Lock() -> Update Counter -> mu.Unlock()
    end
    
    Worker (x50) -->> Main: wg.Done()
    Main->>Main: wg.Wait()
    Main->>Main: Prints summary & elapsed time (15s)
```

> **📌 Note สำหรับคนอ่าน Diagram:** ถ้าใช้เครื่องมือที่รองรับ Markdown ขั้นสูง (เช่น VS Code, Cursor หรือ หน้าเว็บของ GitHub) โค้ด Mermaid ด้านบนจะถูกแปลงเป็นรูปภาพอันสวยงามให้ทันทีครับ
