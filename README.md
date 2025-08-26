# GreenBecak Backend API

Backend API untuk aplikasi GreenBecak menggunakan Go, Gin, GORM, dan MySQL.

## Teknologi yang Digunakan

- **Go 1.21+** - Bahasa pemrograman
- **Gin** - Web framework
- **GORM** - ORM untuk database
- **MySQL** - Database
- **JWT** - Authentication
- **bcrypt** - Password hashing

## Struktur Proyek

```
backend/
├── config/          # Konfigurasi aplikasi
├── database/        # Koneksi dan migrasi database
├── handlers/        # HTTP handlers
├── middleware/      # Middleware (auth, cors, dll)
├── models/          # Model database
├── routes/          # Definisi routes
├── utils/           # Utility functions
├── go.mod           # Dependencies
├── main.go          # Entry point
└── README.md        # Dokumentasi
```

## Setup dan Instalasi

### 1. Prerequisites

- Go 1.21 atau lebih baru
- MySQL 8.0 atau lebih baru
- Git

### 2. Clone dan Setup

```bash
# Clone repository
git clone <repository-url>
cd greenbecak/backend

# Install dependencies
go mod tidy

# Copy environment file
cp env.example .env

# Edit .env file sesuai konfigurasi database Anda
```

### 3. Konfigurasi Database

Edit file `.env`:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=greenbecak_db

# JWT Secret
JWT_SECRET=your-super-secret-jwt-key-here

# Server Configuration
SERVER_PORT=8080
SERVER_MODE=debug

# CORS Configuration
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
```

### 4. Buat Database

```sql
CREATE DATABASE greenbecak_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 5. Jalankan Aplikasi

```bash
# Development mode
go run main.go

# Build dan jalankan
go build -o greenbecak-backend
./greenbecak-backend
```

## API Endpoints

### Authentication

#### POST /api/auth/login
Login user/admin/driver

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
Register customer baru

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

#### POST /api/orders
Buat order baru (Customer)

**Request:**
```json
{
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
Ambil daftar orders (filtered by role)

**Query Parameters:**
- `status`: pending, accepted, completed, cancelled

#### PUT /api/orders/:id
Update status order

**Request:**
```json
{
  "status": "accepted"
}
```

### Tariffs

#### GET /api/tariffs
Ambil daftar tarif aktif

#### GET /api/tariffs/:id
Ambil detail tarif

### Admin Endpoints

#### GET /api/admin/users
Ambil daftar users (Admin only)

#### POST /api/admin/drivers
Buat driver baru (Admin only)

**Request:**
```json
{
  "driver_code": "DRV001",
  "name": "Budi Santoso",
  "phone": "08123456789",
  "email": "budi@example.com",
  "address": "Jl. Malioboro No. 2",
  "id_card": "1234567890123456",
  "license_number": "SIM123456",
  "vehicle_number": "AB1234CD"
}
```

#### GET /api/admin/drivers
Ambil daftar drivers (Admin only)

#### GET /api/admin/analytics
Dashboard analytics (Admin only)

### Driver Endpoints

#### GET /api/driver/orders
Ambil orders untuk driver (Driver only)

#### PUT /api/driver/orders/:id/accept
Accept order (Driver only)

#### PUT /api/driver/orders/:id/complete
Complete order (Driver only)

#### GET /api/driver/earnings
Lihat earnings driver (Driver only)

#### POST /api/driver/withdrawals
Buat withdrawal request (Driver only)

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

## Authentication

API menggunakan JWT token untuk authentication. Setiap request ke protected endpoint harus menyertakan header:

```
Authorization: Bearer <jwt_token>
```

## Role-based Access Control

- **Customer**: Dapat membuat order, melihat order sendiri
- **Driver**: Dapat accept/complete order, lihat earnings, buat withdrawal
- **Admin**: Akses penuh ke semua fitur

## Database Schema

### Users
- id (PK)
- username (unique)
- email (unique)
- password (hashed)
- role (admin/customer)
- name
- phone
- address
- is_active
- timestamps

### Drivers
- id (PK)
- driver_code (unique)
- name
- phone
- email
- address
- id_card
- license_number
- vehicle_number
- status (active/inactive/on_trip)
- rating
- total_trips
- total_earnings
- timestamps

### Orders
- id (PK)
- order_number (unique)
- customer_id (FK)
- driver_id (FK, nullable)
- tariff_id (FK)
- pickup_location
- drop_location
- distance
- price
- status (pending/accepted/completed/cancelled)
- payment_status
- timestamps

### Tariffs
- id (PK)
- name
- description
- min_distance
- max_distance
- base_price
- price_per_km
- is_active
- timestamps

### Payments
- id (PK)
- order_id (FK)
- amount
- method (cash/transfer/qr)
- status (pending/paid/failed/refunded)
- reference
- paid_at
- timestamps

### Withdrawals
- id (PK)
- driver_id (FK)
- amount
- status (pending/approved/rejected/completed)
- bank_name
- account_number
- account_name
- notes
- approved_at
- completed_at
- timestamps

## Error Handling

API mengembalikan error dalam format JSON:

```json
{
  "error": "Error message here"
}
```

Status codes yang digunakan:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

## Development

### Menjalankan Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

### Linting

```bash
golangci-lint run
```

## Deployment

### Build untuk Production

```bash
# Build untuk Linux
GOOS=linux GOARCH=amd64 go build -o greenbecak-backend

# Build untuk Windows
GOOS=windows GOARCH=amd64 go build -o greenbecak-backend.exe

# Build untuk macOS
GOOS=darwin GOARCH=amd64 go build -o greenbecak-backend
```

### Environment Variables untuk Production

```env
SERVER_MODE=release
JWT_SECRET=your-very-secure-jwt-secret
DB_PASSWORD=your-secure-db-password
```

## Contributing

1. Fork repository
2. Buat feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push ke branch (`git push origin feature/amazing-feature`)
5. Buat Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.
