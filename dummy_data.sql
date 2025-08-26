-- =====================================================
-- DUMMY DATA SCRIPT FOR GREENBECAK
-- =====================================================

-- Clear existing data (optional)
-- DELETE FROM payments;
-- DELETE FROM orders;
-- DELETE FROM notifications;
-- DELETE FROM users;

-- =====================================================
-- 1. DUMMY USERS (ADMIN, DRIVER, CUSTOMER)
-- =====================================================

-- Admin Users
INSERT INTO users (name, email, phone, password, role, status, created_at, updated_at) VALUES
('Admin Utama', 'admin@greenbecak.com', '081234567890', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active', NOW(), NOW()),
('Admin Malioboro', 'admin.malioboro@greenbecak.com', '081234567891', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active', NOW(), NOW()),
('Super Admin', 'superadmin@greenbecak.com', '081234567892', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active', NOW(), NOW());

-- Driver Users
INSERT INTO users (name, email, phone, password, role, status, vehicle_number, vehicle_type, license_number, address, emergency_contact, bank_account, bank_name, is_online, created_at, updated_at) VALUES
('Pak Sardi', 'sardi@greenbecak.com', '081234567893', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', 'active', 'B 1234 ABC', 'Becak', 'SIM123456', 'Jl. Malioboro No. 1', '081234567894', '1234567890', 'BCA', true, NOW(), NOW()),
('Pak Joko', 'joko@greenbecak.com', '081234567895', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', 'active', 'B 5678 DEF', 'Becak', 'SIM123457', 'Jl. Malioboro No. 2', '081234567896', '1234567891', 'Mandiri', true, NOW(), NOW()),
('Pak Budi', 'budi@greenbecak.com', '081234567897', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', 'active', 'B 9012 GHI', 'Becak', 'SIM123458', 'Jl. Malioboro No. 3', '081234567898', '1234567892', 'BNI', false, NOW(), NOW()),
('Pak Rudi', 'rudi@greenbecak.com', '081234567899', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', 'active', 'B 3456 JKL', 'Becak', 'SIM123459', 'Jl. Malioboro No. 4', '081234567900', '1234567893', 'BCA', true, NOW(), NOW()),
('Pak Tono', 'tono@greenbecak.com', '081234567901', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'driver', 'inactive', 'B 7890 MNO', 'Becak', 'SIM123460', 'Jl. Malioboro No. 5', '081234567902', '1234567894', 'Mandiri', false, NOW(), NOW());

-- Customer Users
INSERT INTO users (name, email, phone, password, role, status, created_at, updated_at) VALUES
('Budi Santoso', 'budi.santoso@gmail.com', '081234567903', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', 'active', NOW(), NOW()),
('Siti Nurhaliza', 'siti.nurhaliza@gmail.com', '081234567904', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', 'active', NOW(), NOW()),
('Ahmad Rizki', 'ahmad.rizki@gmail.com', '081234567905', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', 'active', NOW(), NOW()),
('Dewi Sartika', 'dewi.sartika@gmail.com', '081234567906', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', 'active', NOW(), NOW()),
('Rizki Pratama', 'rizki.pratama@gmail.com', '081234567907', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', 'active', NOW(), NOW());

-- =====================================================
-- 2. DUMMY TARIFFS
-- =====================================================

INSERT INTO tariffs (name, base_price, price_per_km, description, is_active, created_at, updated_at) VALUES
('Tarif Standar', 10000, 5000, 'Tarif standar untuk perjalanan normal', true, NOW(), NOW()),
('Tarif Malam', 15000, 7000, 'Tarif khusus untuk perjalanan malam (22:00-06:00)', true, NOW(), NOW()),
('Tarif Hujan', 12000, 6000, 'Tarif khusus saat hujan', true, NOW(), NOW()),
('Tarif Promo', 8000, 4000, 'Tarif promo untuk pelanggan baru', true, NOW(), NOW()),
('Tarif VIP', 20000, 8000, 'Tarif premium dengan pelayanan khusus', true, NOW(), NOW());

-- =====================================================
-- 3. DUMMY ORDERS
-- =====================================================

INSERT INTO orders (order_number, customer_id, driver_id, pickup_location, dropoff_location, distance, total_price, status, notes, created_at, updated_at) VALUES
('ORD-2024-001', 6, 1, 'Jl. Malioboro No. 10', 'Tugu Jogja', 2.5, 22500, 'completed', 'Tolong hati-hati ya pak', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
('ORD-2024-002', 7, 2, 'Malioboro Mall', 'Keraton Yogyakarta', 3.0, 25000, 'completed', 'Mau ke keraton', NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY),
('ORD-2024-003', 8, 1, 'Tugu Jogja', 'Jl. Malioboro No. 15', 1.8, 19000, 'in_progress', 'Sedang dalam perjalanan', NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
('ORD-2024-004', 9, 4, 'Keraton Yogyakarta', 'Malioboro Mall', 2.2, 21000, 'pending', 'Menunggu driver', NOW() - INTERVAL 10 MINUTE, NOW() - INTERVAL 10 MINUTE),
('ORD-2024-005', 10, 2, 'Jl. Malioboro No. 20', 'Tugu Jogja', 1.5, 17500, 'accepted', 'Driver sudah menerima order', NOW() - INTERVAL 5 MINUTE, NOW() - INTERVAL 5 MINUTE);

-- =====================================================
-- 4. DUMMY PAYMENTS
-- =====================================================

INSERT INTO payments (order_id, amount, payment_method, status, reference, created_at, updated_at) VALUES
(1, 22500, 'cash', 'completed', 'PAY-2024-001', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
(2, 25000, 'qris', 'completed', 'PAY-2024-002', NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY),
(3, 19000, 'bank_transfer', 'pending', 'PAY-2024-003', NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
(4, 21000, 'e_wallet', 'pending', 'PAY-2024-004', NOW() - INTERVAL 10 MINUTE, NOW() - INTERVAL 10 MINUTE),
(5, 17500, 'cash', 'pending', 'PAY-2024-005', NOW() - INTERVAL 5 MINUTE, NOW() - INTERVAL 5 MINUTE);

-- =====================================================
-- 5. DUMMY NOTIFICATIONS
-- =====================================================

INSERT INTO notifications (user_id, title, message, type, priority, is_read, created_at, updated_at) VALUES
(1, 'Order Baru', 'Ada order baru dari customer Budi Santoso', 'order', 'high', false, NOW() - INTERVAL 5 MINUTE, NOW() - INTERVAL 5 MINUTE),
(2, 'Pembayaran Diterima', 'Pembayaran order ORD-2024-002 telah diterima', 'payment', 'medium', false, NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR),
(6, 'Order Selesai', 'Order ORD-2024-001 telah selesai', 'order', 'low', true, NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
(7, 'Driver Menuju Lokasi', 'Driver Pak Joko sedang menuju lokasi Anda', 'order', 'high', false, NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
(8, 'Pembayaran Berhasil', 'Pembayaran Anda telah berhasil diproses', 'payment', 'medium', false, NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR);

-- =====================================================
-- 6. DUMMY WITHDRAWALS (for drivers)
-- =====================================================

INSERT INTO withdrawals (driver_id, amount, bank_account, bank_name, status, notes, created_at, updated_at) VALUES
(1, 500000, '1234567890', 'BCA', 'completed', 'Penarikan mingguan', NOW() - INTERVAL 7 DAY, NOW() - INTERVAL 7 DAY),
(2, 750000, '1234567891', 'Mandiri', 'pending', 'Penarikan bulanan', NOW() - INTERVAL 2 DAY, NOW() - INTERVAL 2 DAY),
(4, 300000, '1234567893', 'BCA', 'completed', 'Penarikan dana', NOW() - INTERVAL 1 DAY, NOW() - INTERVAL 1 DAY),
(1, 400000, '1234567890', 'BCA', 'pending', 'Penarikan tambahan', NOW() - INTERVAL 12 HOUR, NOW() - INTERVAL 12 HOUR);

-- =====================================================
-- 7. DUMMY LOCATIONS (for drivers)
-- =====================================================

INSERT INTO locations (user_id, latitude, longitude, address, created_at) VALUES
(1, -7.797068, 110.370529, 'Jl. Malioboro No. 1, Yogyakarta', NOW() - INTERVAL 5 MINUTE),
(2, -7.797500, 110.371000, 'Jl. Malioboro No. 2, Yogyakarta', NOW() - INTERVAL 3 MINUTE),
(4, -7.798000, 110.371500, 'Jl. Malioboro No. 4, Yogyakarta', NOW() - INTERVAL 1 MINUTE),
(1, -7.797068, 110.370529, 'Jl. Malioboro No. 1, Yogyakarta', NOW() - INTERVAL 10 MINUTE),
(2, -7.797500, 110.371000, 'Jl. Malioboro No. 2, Yogyakarta', NOW() - INTERVAL 8 MINUTE);

-- =====================================================
-- 8. DUMMY ALERTS (for monitoring)
-- =====================================================

INSERT INTO alerts (title, message, severity, is_active, created_at, updated_at) VALUES
('Server Performance Warning', 'CPU usage mencapai 80%', 'warning', true, NOW() - INTERVAL 1 HOUR, NOW() - INTERVAL 1 HOUR),
('Database Connection Alert', 'Koneksi database lambat', 'error', true, NOW() - INTERVAL 30 MINUTE, NOW() - INTERVAL 30 MINUTE),
('Payment Gateway Issue', 'Payment gateway tidak merespons', 'critical', true, NOW() - INTERVAL 15 MINUTE, NOW() - INTERVAL 15 MINUTE),
('Driver Offline Alert', '5 driver offline dalam 1 jam terakhir', 'warning', true, NOW() - INTERVAL 2 HOUR, NOW() - INTERVAL 2 HOUR);

-- =====================================================
-- 9. UPDATE USER LOCATIONS (for current location)
-- =====================================================

UPDATE users SET 
    current_location_latitude = -7.797068,
    current_location_longitude = 110.370529
WHERE id = 1;

UPDATE users SET 
    current_location_latitude = -7.797500,
    current_location_longitude = 110.371000
WHERE id = 2;

UPDATE users SET 
    current_location_latitude = -7.798000,
    current_location_longitude = 110.371500
WHERE id = 4;

-- =====================================================
-- SUMMARY
-- =====================================================

-- Total records created:
-- - 3 Admin users
-- - 5 Driver users  
-- - 5 Customer users
-- - 5 Tariffs
-- - 5 Orders
-- - 5 Payments
-- - 5 Notifications
-- - 4 Withdrawals
-- - 5 Location records
-- - 4 Alerts

-- Login credentials for testing:
-- Admin: admin@greenbecak.com / password
-- Driver: sardi@greenbecak.com / password
-- Customer: budi.santoso@gmail.com / password
