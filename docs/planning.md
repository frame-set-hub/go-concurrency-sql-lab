# Planning & Roadmap

โปรเจกต์นี้ออกแบบมาเพื่อฝึกฝน **Goroutines** และ **SQL (PostgreSQL)** ตั้งแต่ระดับพื้นฐานไปจนถึงขั้นสูง โดยจะจำลองระบบ **E-Commerce Order Processing & Analytics** ที่ต้องรับมือกับข้อมูลจำนวนมากในเวลาอันรวดเร็ว

## 🗄️ Phase 1: SQL Analytics (Basic to Master)
เริ่มจากการออกแบบ Database และเขียน SQL วิเคราะห์ข้อมูล โดยเน้นที่โจทย์จาก `GROUP BY` และ `HAVING` ตามที่คุณต้องการ

1. **Schema Design**
   - `users` (id, name, created_at)
   - `products` (id, name, category, price)
   - `orders` (id, user_id, status, created_at)
   - `order_items` (id, order_id, product_id, quantity, unit_price)
2. **Basic SQL**
   - การ JOIN Tables (`orders` + `users` + `order_items`)
3. **Intermediate SQL (Targeted: GROUP BY & HAVING)**
   - 📊 **Top Spenders:** หา User ที่มียอดใช้จ่ายรวมมากกว่า 10,000 บาท (`GROUP BY user_id HAVING SUM(price) > 10000`)
   - 📦 **Best Sellers by Category:** หาสินค้าที่ขายได้มากกว่า 500 ชิ้นในแต่ละหมวดหมู่
   - 🕒 **Peak Hours:** หาช่วงเวลา (ชั่วโมง) ที่มีคนสั่งซื้อมากที่สุด และมีจำนวนออเดอร์เฉลี่ยเกินเกณฑ์ที่กำหนด
4. **Master SQL (Bonus)**
   - Window Functions: คำนวณยอดขายสะสม (Running Total)
   - CTEs (Common Table Expressions): เขียน Query ที่ซับซ้อนให้อ่านง่ายขึ้น
   - การทำ Index เพื่อให้ Query ที่มี `GROUP BY` ทำงานได้เร็วขึ้น

## 🐹 Phase 2: Go Routine & Concurrency (Basic)
เรียนรู้กลไกการทำงานของ Goroutines สเต็ปบายสเต็ป

1. **Goroutines & WaitGroups (`sync.WaitGroup`)**
   - จำลองการดึง order จาก API 10 ตัวพร้อมๆ กัน โดยไม่ block main thread
2. **Channels (Data passing)**
   - สร้าง Generator ส่งข้อมูล order ลง Channel เพื่อให้ Processor รอรับไปทำงาน
3. **Data Races & Mutex (`sync.Mutex`)**
   - ฝึกแก้ปัญหา Data Race เมื่อหลายๆ Goroutine พยายามจะ update ตัวแปรเดียวกัน (เช่น ตัวนับยอดขายรวมใน memory)

## 🚀 Phase 3: The Integration (Intermediate)
จับ Go กับ PostgreSQL มาทำงานร่วมกันภายใต้ Concurrency

1. **Concurrent Data Seeding (Bulk Insert)**
   - **โจทย์:** เราต้องการ dummy data 1,000,000 orders เพื่อมาเทสสคริปต์ SQL ของเรา
   - **วิธีทำ:** ใช้ Goroutines หลายๆ ตัวสร้างข้อมูลและ Insert ลง Postgres เป็น Batch (ใช้ `pgx` หรือ `database/sql`)
2. **Connection Pooling**
   - การจูน `SetMaxOpenConns` และ `SetMaxIdleConns` เพื่อไม่ให้ Goroutines แห่กันไปยิง Database จนพัง (Too many clients)

## 🧠 Phase 4: Master Go Concurrency Patterns
นำ Pattern จริงที่ใช้ใน Production มาประยุกต์ใช้

1. **Worker Pool Pattern 🌟**
   - แทนที่จะเปิด Goroutine ล้านตัวเพื่อ insert ล้าน orders เราจะใช้ Worker แค่ 50-100 ตัว คอยรอรับงานจาก Channel (ลดภาระ Memory)
2. **Fan-Out / Fan-In**
   - รัน SQL Query ย่อยๆ 3 ตัวพร้อมกัน (Fan-Out) เช่น หายอดขาย, หา users ใหม่, หาสินค้าขายดี แล้วนำผลลัพธ์มารวมกัน (Fan-In) เพื่อส่งเป็น API Response เดียว
3. **Context & Graceful Shutdown (`context.Context`)**
   - หาก Query รันนานเกินไป หรือมี User กด Cancel (Ctrl+C) ระบบจะต้องหยุด Worker ทุกตัวอย่างปลอดภัย

## 🏆 Phase 5: The Grand Finale
สร้าง Service จริงขึ้นมา 1 ตัว ที่รวบรวมทุกอย่าง:

- **Background Job:** ตั้ง Ticker รันตัว Worker ทุกๆ 5 นาทีเพื่อใช้ SQL (`GROUP BY`, `HAVING`) ทำการสรุปยอดขายไปเก็บใน Table `daily_summaries`
- **Real-time API:** สร้าง Web Server ยิงเข้ามาดูผลลัพธ์ โดยข้อมูลจะถูกประมวลผลล่วงหน้ามาแล้วด้วย Concurrent Job
