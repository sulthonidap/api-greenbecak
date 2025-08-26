# GreenBecak API Documentation

## Overview

GreenBecak API adalah RESTful API untuk aplikasi manajemen becak yang dibangun dengan Go, Gin, GORM, dan MySQL.

## Base URL

```
http://localhost:8080
```

## Authentication

API menggunakan JWT (JSON Web Token) untuk authentication. Setiap request ke protected endpoint harus menyertakan header:

```
Authorization: Bearer <jwt_token>
```

## Response Format

Semua response menggunakan format JSON:

### Success Response
```json
{
  "message": "Success message",
  "data": { ... }
}
```

### Error Response
```json
{
  "error": "Error message"
}
```

## Status Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `422` - Unprocessable Entity
- `429` - Too Many Requests
- `500` - Internal Server Error
- `503` - Service Unavailable

## Rate Limiting

- **Auth endpoints**: 10 requests per minute
- **Other endpoints**: 100 requests per minute

## Endpoints

### Location Tracking

#### POST /api/driver/location
Update lokasi driver secara real-time.

**Request Body:**
```json
{
  "latitude": -7.797068,
  "longitude": 110.370529,
  "accuracy": 10.5,
  "speed": 25.0,
  "heading": 90.0,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Response:**
```json
{
  "message": "Location updated successfully",
  "location": {
    "driver_id": 1,
    "latitude": -7.797068,
    "longitude": 110.370529,
    "accuracy": 10.5,
    "speed": 25.0,
    "heading": 90.0,
    "is_online": true,
    "last_seen": "2024-01-01T12:00:00Z"
  }
}
```

#### GET /api/location/drivers/nearby
Mendapatkan driver yang berada di sekitar lokasi.

**Query Parameters:**
- `lat` (required): Latitude
- `lng` (required): Longitude  
- `radius` (optional): Radius dalam km (default: 5)
- `limit` (optional): Jumlah driver maksimal (default: 10)

**Response:**
```json
{
  "drivers": [
    {
      "driver_id": 1,
      "latitude": -7.797068,
      "longitude": 110.370529,
      "distance": 0.5,
      "is_online": true,
      "last_seen": "2024-01-01T12:00:00Z"
    }
  ],
  "count": 1,
  "radius": 5
}
```

#### GET /api/location/routes/:order_id
Mendapatkan rute driver untuk order tertentu.

**Response:**
```json
{
  "order_id": "1",
  "driver_location": {
    "latitude": -7.797068,
    "longitude": 110.370529
  },
  "pickup_location": "Malioboro Mall",
  "drop_location": "Hotel Indonesia",
  "route": {
    "waypoints": [
      {"lat": -7.797068, "lng": 110.370529},
      {"lat": -7.798068, "lng": 110.371529},
      {"lat": -7.799068, "lng": 110.372529}
    ],
    "distance": 2.5,
    "estimated_time": 15
  },
  "estimated_time": 15,
  "distance": 2.5
}
```

### Payment Management

#### POST /api/payments
Membuat pembayaran baru.

**Request Body:**
```json
{
  "order_id": 1,
  "method": "cash",
  "amount": 25000,
  "notes": "Pembayaran tunai"
}
```

**Response:**
```json
{
  "message": "Payment created successfully",
  "payment": {
    "id": 1,
    "order_id": 1,
    "amount": 25000,
    "method": "cash",
    "status": "pending",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### GET /api/payments
Mendapatkan daftar pembayaran.

**Query Parameters:**
- `page` (optional): Halaman (default: 1)
- `limit` (optional): Jumlah item per halaman (default: 10)

#### PUT /api/payments/:id/status
Update status pembayaran.

**Request Body:**
```json
{
  "status": "paid"
}
```

#### POST /api/payments/:id/process
Memproses pembayaran (simulasi).

#### GET /api/payments/stats
Mendapatkan statistik pembayaran.

### Notification System

#### POST /api/notifications
Membuat notifikasi baru (admin only).

**Request Body:**
```json
{
  "user_id": 1,
  "title": "Order Baru",
  "message": "Anda memiliki order baru",
  "type": "order",
  "priority": "high",
  "data": {
    "order_id": 1,
    "amount": 25000
  }
}
```

#### GET /api/notifications
Mendapatkan daftar notifikasi user.

**Query Parameters:**
- `page` (optional): Halaman (default: 1)
- `limit` (optional): Jumlah item per halaman (default: 20)
- `type` (optional): Filter berdasarkan tipe
- `read` (optional): Filter berdasarkan status baca (true/false)
- `priority` (optional): Filter berdasarkan prioritas

#### PUT /api/notifications/:id/read
Menandai notifikasi sebagai sudah dibaca.

#### PUT /api/notifications/read-all
Menandai semua notifikasi sebagai sudah dibaca.

#### DELETE /api/notifications/:id
Menghapus notifikasi.

#### GET /api/notifications/stats
Mendapatkan statistik notifikasi.

#### POST /api/admin/notifications/bulk
Mengirim notifikasi ke multiple users (admin only).

**Request Body:**
```json
{
  "user_ids": [1, 2, 3],
  "title": "Promo Spesial",
  "message": "Dapatkan diskon 20%",
  "type": "promo",
  "priority": "normal"
}
```

### System Endpoints

#### GET /health
Health check endpoint dengan detail lengkap.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "2h30m15s",
  "version": "1.0.0",
  "services": {
    "database": "healthy",
    "api": "healthy",
    "memory": "healthy",
    "disk": "healthy"
  }
}
```

#### GET /ready
Readiness check untuk Kubernetes.

#### GET /live
Liveness check untuk Kubernetes.

#### GET /metrics
Metrics endpoint untuk monitoring.

#### GET /alerts
Mendapatkan semua alerts.

#### GET /alerts/active
Mendapatkan active alerts saja.

### Authentication

#### POST /api/auth/login
Login untuk user/admin/driver.

**Request:**
```json
{
  "username": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "jwt_token_here",
  "user": {
    "id": 1,
    "username": "user@example.com",
    "name": "John Doe",
    "role": "customer"
  },
  "message": "Login successful"
}
```

#### POST /api/auth/register
Register customer baru.

**Request:**
```json
{
  "username": "newuser",
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe",
  "phone": "08123456789",
  "address": "Jl. Malioboro No. 1"
}
```

### Orders

#### POST /api/orders/public
Buat order baru untuk customer tanpa akun (menggunakan kode becak dari sticker).

**Request:**
```json
{
  "becak_code": "DRV-001",
  "tariff_id": 1,
  "customer_phone": "08123456789",
  "customer_name": "Budi Santoso",
  "notes": "Tolong hati-hati ya pak"
}
```

**Response:**
```json
{
  "message": "Order created successfully",
  "order": {
    "id": 1,
    "order_number": "ORD-1703123456",
    "becak_code": "DRV-001",
    "price": 10000,
    "status": "pending",
    "customer_phone": "08123456789",
    "customer_name": "Budi Santoso"
  },
  "driver": {
    "name": "Pak Seno",
    "phone": "08123456789"
  }
}
```

#### GET /api/orders/history
Ambil riwayat order berdasarkan nomor telepon customer (public).

**Query Parameters:**
- `phone` (required): Nomor telepon customer
- `page`: page number (default: 1)
- `limit`: items per page (default: 10)

**Response:**
```json
{
  "orders": [
    {
      "id": 1,
      "order_number": "ORD-1703123456",
      "pickup_location": "Jl. Malioboro No. 10",
      "drop_location": "Tugu Jogja",
      "price": 10000,
      "status": "completed",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

#### POST /api/orders
Buat order baru (Customer dengan akun).

**Request:**
```json
{
  "customer_id": 1,
  "tariff_id": 1,
  "pickup_location": "Malioboro",
  "drop_location": "Tugu Jogja",
  "distance": 2.5,
  "customer_phone": "08123456789",
  "customer_name": "John Doe",
  "notes": "Tolong hati-hati"
}
```

#### GET /api/orders
Ambil daftar orders (filtered by role).

**Query Parameters:**
- `status`: pending, accepted, completed, cancelled
- `page`: page number (default: 1)
- `limit`: items per page (default: 10)

#### GET /api/orders/:id
Ambil detail order berdasarkan ID.

#### PUT /api/orders/:id
Update status order.

**Request:**
```json
{
  "status": "accepted"
}
```

#### PUT /api/orders/:id/location
Update lokasi pickup dan drop order (Driver).

**Request:**
```json
{
  "pickup_location": "Jl. Malioboro No. 10",
  "drop_location": "Tugu Jogja",
  "distance": 2.5
}
```

### Tariffs

#### GET /api/tariffs
Ambil daftar tarif aktif.

**Query Parameters:**
- `is_active`: true/false

#### GET /api/tariffs/:id
Ambil detail tarif.

### Admin Endpoints

#### POST /api/admin/users
Buat user baru (Admin only).

**Request:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "password123",
  "name": "John Doe",
  "phone": "08123456789",
  "address": "Jl. Malioboro No. 1",
  "role": "customer"
}
```

**Untuk membuat driver:**
```json
{
  "username": "DRV-001",
  "email": "seno@greenbecak.com",
  "password": "password123",
  "name": "Pak Seno",
  "phone": "08123456789",
  "address": "Jl. Malioboro No. 10",
  "role": "driver",
  "driver_code": "DRV-001",
  "id_card": "1234567890123456",
  "vehicle_number": "AB 1234 XX"
}
```

**Note:** Jika role = "driver", otomatis akan membuat record driver juga.

#### GET /api/admin/users
Ambil daftar users (Admin only).

**Query Parameters:**
- `role`: admin, customer, driver
- `is_active`: true/false

#### GET /api/admin/users/:id
Ambil detail user (Admin only).

#### PUT /api/admin/users/:id
Update user (Admin only).

#### DELETE /api/admin/users/:id
Delete user (Admin only).

#### POST /api/admin/drivers
Buat driver baru (Admin only).

**Request:**
```json
{
  "driver_code": "DRV001",
  "name": "Budi Santoso",
  "phone": "08123456789",
  "address": "Jl. Malioboro No. 2",
  "id_card": "1234567890123456",
  "vehicle_number": "AB1234CD"
}
```

**Note:** Email akan otomatis dibuat dengan format `driver_code@drivers.local`

#### GET /api/admin/drivers
Ambil daftar drivers (Admin only).

**Query Parameters:**
- `status`: active, inactive, on_trip
- `is_active`: true/false

#### GET /api/admin/drivers/:id
Ambil detail driver (Admin only).

#### PUT /api/admin/drivers/:id
Update driver (Admin only).

#### DELETE /api/admin/drivers/:id
Delete driver (Admin only).

#### GET /api/admin/drivers/:id/performance
Ambil performance driver (Admin only).

#### POST /api/admin/tariffs
Buat tarif baru (Admin only).

**Request:**
```json
{
  "name": "Dekat",
  "description": "Jarak dekat (< 3 km)",
  "min_distance": 0,
  "max_distance": 3,
  "base_price": 10000,
  "price_per_km": 0
}
```

#### PUT /api/admin/tariffs/:id
Update tarif (Admin only).

#### PUT /api/admin/tariffs/:id/active
Aktifkan/nonaktifkan tarif (Admin only).

**Request:**
```json
{
  "is_active": true
}
```

#### DELETE /api/admin/tariffs/:id
Delete tarif (Admin only).

#### GET /api/admin/analytics
Dashboard analytics (Admin only).

#### GET /api/admin/analytics/revenue
Revenue analytics (Admin only).

**Query Parameters:**
- `period`: week, month

#### GET /api/admin/analytics/orders
Order analytics (Admin only).

**Query Parameters:**
- `period`: week, month

#### GET /api/admin/withdrawals
Ambil daftar withdrawal requests (Admin only).

**Query Parameters:**
- `status`: pending, approved, rejected, completed
- `driver_id`: driver ID

#### GET /api/admin/withdrawals/:id
Ambil detail withdrawal (Admin only).

#### PUT /api/admin/withdrawals/:id
Update withdrawal status (Admin only).

**Request:**
```json
{
  "status": "approved",
  "notes": "Approved by admin"
}
```

### Driver Endpoints

#### GET /api/driver/orders
Ambil orders untuk driver (Driver only).

**Query Parameters:**
- `status`: pending, accepted, completed, cancelled

#### PUT /api/driver/orders/:id/accept
Accept order (Driver only).

#### PUT /api/driver/orders/:id/complete
Complete order (Driver only).

#### GET /api/driver/earnings
Lihat earnings driver (Driver only).

#### POST /api/driver/withdrawals
Buat withdrawal request (Driver only).

**Request:**
```json
{
  "amount": 500000,
  "bank_name": "BCA",
  "account_number": "1234567890",
  "account_name": "Budi Santoso",
  "notes": "Penarikan untuk kebutuhan keluarga"
}
```

#### GET /api/driver/withdrawals
Lihat withdrawal history driver (Driver only).

**Query Parameters:**
- `status`: pending, approved, rejected, completed

## Error Handling

API mengembalikan error dalam format yang konsisten:

```json
{
  "error": "Error message here"
}
```

### Common Error Messages

- `"Authorization header required"` - Token tidak disertakan
- `"Invalid token"` - Token tidak valid atau expired
- `"Admin access required"` - Endpoint memerlukan akses admin
- `"Driver access required"` - Endpoint memerlukan akses driver
- `"Rate limit exceeded"` - Terlalu banyak request
- `"Content-Type must be application/json"` - Content-Type salah

## Pagination

Untuk endpoint yang mendukung pagination, response akan memiliki format:

```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "pages": 10
  }
}
```

## Filtering

Banyak endpoint mendukung filtering melalui query parameters:

- `status`: Filter berdasarkan status
- `is_active`: Filter berdasarkan status aktif
- `role`: Filter berdasarkan role user
- `driver_id`: Filter berdasarkan driver ID

## Sorting

Endpoint yang mendukung sorting akan menggunakan query parameter `sort`:

```
GET /api/orders?sort=created_at:desc
GET /api/drivers?sort=name:asc
```

## Search

Endpoint yang mendukung search akan menggunakan query parameter `search`:

```
GET /api/users?search=john
GET /api/drivers?search=budi
```

## WebSocket Support

Untuk real-time updates, API mendukung WebSocket connections:

```
ws://localhost:8080/ws
```

### WebSocket Events

- `order.created` - Order baru dibuat
- `order.updated` - Status order berubah
- `driver.location` - Update lokasi driver
- `alert.new` - Alert baru

## Testing

API menyediakan endpoint untuk testing:

### GET /test/db
Test database connection.

### GET /test/auth
Test authentication.

### POST /test/cleanup
Cleanup test data.

## Development

### Local Development

1. Clone repository
2. Copy `env.example` ke `.env`
3. Setup database
4. Run `go mod tidy`
5. Run `go run main.go`

### Environment Variables

```env
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=greenbecak_db

# JWT
JWT_SECRET=your-super-secret-jwt-key-here

# Server
SERVER_PORT=8080
SERVER_MODE=debug

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

### Database Setup

```sql
CREATE DATABASE greenbecak_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### Seeding Data

```bash
go run scripts/seed.go
```

## Production Deployment

### Docker

```bash
docker build -t greenbecak-backend .
docker run -p 8080:8080 --env-file .env greenbecak-backend
```

### Docker Compose

```bash
docker-compose up -d
```

### Environment Variables for Production

```env
SERVER_MODE=release
JWT_SECRET=your-very-secure-jwt-secret
DB_PASSWORD=your-secure-db-password
```

## Monitoring

### Health Checks

- `/health` - Comprehensive health check
- `/ready` - Readiness check
- `/live` - Liveness check

### Metrics

- `/metrics` - Application metrics
- `/alerts` - System alerts

### Logging

API menggunakan structured logging dengan level:
- `DEBUG` - Development mode
- `INFO` - General information
- `WARN` - Warnings
- `ERROR` - Errors

## Security

### Authentication
- JWT tokens dengan expiration 24 jam
- Password hashing dengan bcrypt
- Rate limiting untuk mencegah brute force

### Authorization
- Role-based access control (RBAC)
- Admin, Driver, Customer roles
- Endpoint-specific permissions

### Data Protection
- Input validation
- SQL injection prevention (GORM)
- XSS protection
- CORS configuration

## Performance

### Database
- Connection pooling
- Indexes untuk query optimization
- Prepared statements

### Caching
- Response caching untuk static data
- Database query caching

### Optimization
- Gzip compression
- Response size optimization
- Efficient JSON serialization

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check database credentials
   - Ensure MySQL is running
   - Verify network connectivity

2. **JWT Token Invalid**
   - Check JWT_SECRET environment variable
   - Verify token expiration
   - Ensure proper token format

3. **Rate Limit Exceeded**
   - Wait before making more requests
   - Implement proper request throttling

4. **CORS Errors**
   - Check CORS_ALLOWED_ORIGINS configuration
   - Verify frontend origin

### Debug Mode

Set `SERVER_MODE=debug` untuk mendapatkan detailed logs:

```env
SERVER_MODE=debug
```

### Logs

Check application logs for detailed error information:

```bash
tail -f logs/app.log
```

## Support

Untuk bantuan teknis, hubungi:
- Email: support@greenbecak.com
- Documentation: https://docs.greenbecak.com
- GitHub Issues: https://github.com/greenbecak/backend/issues
