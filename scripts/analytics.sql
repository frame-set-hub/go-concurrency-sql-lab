-- 1️⃣ Top Spenders: หา User สายเปย์ ที่มียอดใช้จ่ายรวมเกิน 10,000 บาท
SELECT 
    u.id, 
    u.name, 
    SUM(oi.quantity * oi.unit_price) AS total_spent
FROM users u
JOIN orders o ON u.id = o.user_id
JOIN order_items oi ON o.id = oi.order_id
GROUP BY u.id, u.name
HAVING SUM(oi.quantity * oi.unit_price) > 10000
ORDER BY total_spent DESC
LIMIT 10;

-- 2️⃣ Best Sellers by Category: หาสินค้ายอดฮิตขายดีที่สุดประจำหมวดหมู่
-- (ใช้ CTE และ Window Function ช่วยในการ Rank อันดับ 1 ของแต่ละ Category)
WITH CategorySales AS (
    SELECT p.category, p.id, p.name, SUM(oi.quantity) as total_sold
    FROM products p
    JOIN order_items oi ON p.id = oi.product_id
    GROUP BY p.category, p.id, p.name
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

-- 3️⃣ Peak Hours: หาชั่วโมงทองคำ (Peak Hour) ที่มีคนซื้อของมากที่สุด
SELECT 
    EXTRACT(HOUR FROM created_at) AS order_hour, 
    COUNT(id) AS total_orders
FROM orders
GROUP BY EXTRACT(HOUR FROM created_at)
ORDER BY total_orders DESC
LIMIT 5;
