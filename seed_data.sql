-- =====================================================
-- DUMMY DATA SCRIPT FOR GREENBECAK
-- =====================================================

-- Clear existing data (optional)
-- DELETE FROM payments;
-- DELETE FROM orders;
-- DELETE FROM notifications;
-- DELETE FROM withdrawals;
-- DELETE FROM drivers;
-- DELETE FROM users;

-- =====================================================
-- 1. DUMMY USERS (ADMIN, CUSTOMER, DRIVER)
-- =====================================================

-- Admin Users
INSERT INTO users (username, name, email, phone, password, role, is_active, created_at, updated_at) VALUES
('admin_utama', 'Admin Utama', 'admin@greenbecak.com', '081234567890', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', true, NOW(), NOW()),
('admin_malioboro', 'Admin Malioboro', 'admin.malioboro@greenbecak.com', '081234567891', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', true, NOW(), NOW()),
('super_admin', 'Super Admin', 'superadmin@greenbecak.com', '081234567892', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', true, NOW(), NOW());

-- Customer Users
INSERT INTO users (username, name, email, phone, password, role, is_active, created_at, updated_at) VALUES
('budi_santoso', 'Budi Santoso', 'budi.santoso@gmail.com', '081234567903', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', true, NOW(), NOW()),
('siti_nurhaliza', 'Siti Nurhaliza', 'siti.nurhaliza@gmail.com', '081234567904', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', true, NOW(), NOW()),
('ahmad_rizki', 'Ahmad Rizki', 'ahmad.rizki@gmail.com', '081234567905', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', true, NOW(), NOW()),
('dewi_sartika', 'Dewi Sartika', 'dewi.sartika@gmail.com', '081234567906', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', true, NOW(), NOW()),
('rizki_pratama', 'Rizki Pratama', 'rizki.pratama@gmail.com', '081234567907', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', true, NOW(), NOW());

-- Driver Users (akan terhubung dengan tabel drivers)
INSERT INTO users (username, name, email, phone, password, role, is_active, created_at, updated_at) VALUES
('driver_seno', 'Pak Seno', 'driver.seno@greenbecak.com', '08123456789', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', true, NOW(), NOW()),
('driver_joko', 'Pak Joko', 'driver.joko@greenbecak.com', '08123456790', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', true, NOW(), NOW()),
('driver_sari', 'Pak Sari', 'driver.sari@greenbecak.com', '08123456791', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', false, NOW(), NOW()),
('driver_rudi', 'Pak Rudi', 'driver.rudi@greenbecak.com', '08123456792', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', true, NOW(), NOW()),
('driver_bambang', 'Pak Bambang', 'driver.bambang@greenbecak.com', '08123456793', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', true, NOW(), NOW());

-- =====================================================
-- 2. DUMMY DRIVERS (terhubung dengan users)
-- =====================================================

-- Create drivers dengan relasi ke users dan vehicle_type
INSERT INTO drivers (user_id, driver_code, name, phone, email, address, id_card, vehicle_number, vehicle_type, status, is_active, rating, total_trips, total_earnings, created_at, updated_at) VALUES
(8, 'DRV-001', 'Pak Seno', '08123456789', 'driver.seno@greenbecak.com', 'Jl. Malioboro No. 10', '1234567890123456', 'AB 1234 XX', 'becak_manual', 'active', true, 4.5, 150, 2500000, NOW(), NOW()),
(9, 'DRV-002', 'Pak Joko', '08123456790', 'driver.joko@greenbecak.com', 'Jl. Malioboro No. 11', '1234567890123457', 'AB 1235 XX', 'becak_motor', 'active', true, 4.8, 200, 3000000, NOW(), NOW()),
(10, 'DRV-003', 'Pak Sari', '08123456791', 'driver.sari@greenbecak.com', 'Jl. Malioboro No. 12', '1234567890123458', 'AB 1236 XX', 'becak_listrik', 'inactive', false, 4.2, 100, 1500000, NOW(), NOW()),
(11, 'DRV-004', 'Pak Rudi', '08123456792', 'driver.rudi@greenbecak.com', 'Jl. Malioboro No. 13', '1234567890123459', 'AB 1237 XX', 'andong', 'on_trip', true, 4.7, 180, 2800000, NOW(), NOW()),
(12, 'DRV-005', 'Pak Bambang', '08123456793', 'driver.bambang@greenbecak.com', 'Jl. Malioboro No. 14', '1234567890123460', 'AB 1238 XX', 'becak_motor', 'active', true, 4.6, 220, 3200000, NOW(), NOW());

-- =====================================================
-- 3. DUMMY TARIFFS (flat pricing)
-- =====================================================

INSERT INTO tariffs (name, min_distance, max_distance, price, destinations, is_active, created_at, updated_at) VALUES
('Dekat', 0, 3, 10000, 'Benteng Vredeburg, Bank Indonesia, Malioboro Mall', true, NOW(), NOW()),
('Sedang', 3, 7, 20000, 'Taman Sari, Alun-Alun Selatan, Keraton Yogyakarta', true, NOW(), NOW()),
('Jauh', 7, 15, 30000, 'Tugu Jogja, Stasiun Lempuyangan, Bandara Adisucipto', true, NOW(), NOW()),
('Sangat Jauh', 15, 25, 40000, 'Candi Prambanan, Candi Borobudur, Gunung Merapi', true, NOW(), NOW()),
('Tarif Malam', 0, 10, 25000, 'Semua destinasi (22:00-06:00)', true, NOW(), NOW()),
('Tarif Hujan', 0, 10, 20000, 'Semua destinasi saat hujan', true, NOW(), NOW()),
('Tarif Promo', 0, 5, 8000, 'Destinasi terbatas untuk pelanggan baru', true, NOW(), NOW()),
('Tarif VIP', 0, 20, 50000, 'Semua destinasi dengan pelayanan premium', true, NOW(), NOW());

-- =====================================================
-- 4. DUMMY ORDERS
-- =====================================================

INSERT INTO orders (order_number, customer_id, driver_id, tariff_id, pickup_location, drop_location, distance, price, status, payment_status, customer_phone, customer_name, notes, created_at, updated_at) VALUES
('ORD-2024-001', 4, 1, 1, 'Jl. Malioboro No. 10', 'Benteng Vredeburg', 2.5, 10000, 'completed', 'completed', '081234567903', 'Budi Santoso', 'Tolong hati-hati ya pak', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
('ORD-2024-002', 5, 2, 2, 'Malioboro Mall', 'Keraton Yogyakarta', 4.0, 20000, 'completed', 'completed', '081234567904', 'Siti Nurhaliza', 'Mau ke keraton', NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY),
('ORD-2024-003', 6, 1, 1, 'Tugu Jogja', 'Jl. Malioboro No. 15', 1.8, 10000, 'accepted', 'pending', '081234567905', 'Ahmad Rizki', 'Sedang dalam perjalanan', NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
('ORD-2024-004', 7, 4, 3, 'Keraton Yogyakarta', 'Bandara Adisucipto', 8.5, 30000, 'pending', 'pending', '081234567906', 'Dewi Sartika', 'Menunggu driver', NOW() - INTERVAL 10 MINUTE, NOW() - INTERVAL 10 MINUTE),
('ORD-2024-005', 8, 2, 2, 'Jl. Malioboro No. 20', 'Taman Sari', 5.2, 20000, 'accepted', 'pending', '081234567907', 'Rizki Pratama', 'Driver sudah menerima order', NOW() - INTERVAL 5 MINUTE, NOW() - INTERVAL 5 MINUTE);

-- =====================================================
-- 5. DUMMY PAYMENTS
-- =====================================================

INSERT INTO payments (order_id, amount, method, status, reference, created_at, updated_at) VALUES
(1, 10000, 'cash', 'paid', 'PAY-2024-001', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
(2, 20000, 'qr', 'paid', 'PAY-2024-002', NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY),
(3, 10000, 'transfer', 'pending', 'PAY-2024-003', NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
(4, 30000, 'qr', 'pending', 'PAY-2024-004', NOW() - INTERVAL 10 MINUTE, NOW() - INTERVAL 10 MINUTE),
(5, 20000, 'cash', 'pending', 'PAY-2024-005', NOW() - INTERVAL 5 MINUTE, NOW() - INTERVAL 5 MINUTE);

-- =====================================================
-- 6. DUMMY NOTIFICATIONS
-- =====================================================

INSERT INTO notifications (user_id, title, message, type, priority, is_read, created_at, updated_at) VALUES
(1, 'Order Baru', 'Ada order baru dari customer Budi Santoso', 'order', 'high', false, NOW() - INTERVAL 5 MINUTE, NOW() - INTERVAL 5 MINUTE),
(2, 'Pembayaran Diterima', 'Pembayaran order ORD-2024-002 telah diterima', 'payment', 'medium', false, NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR),
(4, 'Order Selesai', 'Order ORD-2024-001 telah selesai', 'order', 'low', true, NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
(5, 'Driver Menuju Lokasi', 'Driver Pak Joko sedang menuju lokasi Anda', 'order', 'high', false, NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
(6, 'Pembayaran Berhasil', 'Pembayaran Anda telah berhasil diproses', 'payment', 'medium', false, NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR),
(8, 'Order Diterima', 'Order ORD-2024-003 telah diterima', 'order', 'high', false, NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
(9, 'Pembayaran Cash', 'Pembayaran cash untuk order ORD-2024-002', 'payment', 'medium', false, NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY);

-- =====================================================
-- 7. DUMMY WITHDRAWALS (for drivers)
-- =====================================================

INSERT INTO withdrawals (driver_id, amount, bank_name, account_number, account_name, status, notes, created_at, updated_at) VALUES
(1, 500000, 'BCA', '1234567890', 'Pak Seno', 'completed', 'Penarikan mingguan', NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 7 DAY),
(2, 750000, 'Mandiri', '1234567891', 'Pak Joko', 'pending', 'Penarikan bulanan', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
(4, 300000, 'BCA', '1234567893', 'Pak Rudi', 'completed', 'Penarikan dana', NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY),
(1, 400000, 'BCA', '1234567890', 'Pak Seno', 'pending', 'Penarikan tambahan', NOW() - INTERVAL 12 HOUR, NOW() - INTERVAL 12 HOUR);

-- =====================================================
-- SUMMARY
-- =====================================================

-- Total records created:
-- - 3 Admin users
-- - 5 Customer users  
-- - 5 Driver users (dengan relasi ke tabel drivers)
-- - 5 Driver records (terhubung dengan driver users)
-- - 8 Tariffs (flat pricing berdasarkan jarak)
-- - 5 Orders
-- - 5 Payments
-- - 7 Notifications
-- - 4 Withdrawals

-- Login credentials for testing:
-- Admin: admin@greenbecak.com / password
-- Driver: driver.seno@greenbecak.com / password
-- Customer: budi.santoso@gmail.com / password

-- Tariff System (Flat Pricing):
-- - Dekat (0-3 km): Rp 10.000
-- - Sedang (3-7 km): Rp 20.000
-- - Jauh (7-15 km): Rp 30.000
-- - Sangat Jauh (15-25 km): Rp 40.000
-- - Tarif Malam (0-10 km): Rp 25.000
-- - Tarif Hujan (0-10 km): Rp 20.000
-- - Tarif Promo (0-5 km): Rp 8.000
-- - Tarif VIP (0-20 km): Rp 50.000
