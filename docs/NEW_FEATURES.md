# Fitur API Baru - GreenBecak

## Overview

Dokumen ini menjelaskan fitur-fitur API baru yang telah ditambahkan ke GreenBecak untuk meningkatkan fungsionalitas aplikasi.

## 1. Payment Management System

### Fitur Utama
- ✅ Pembuatan pembayaran baru
- ✅ Tracking status pembayaran
- ✅ Processing pembayaran (simulasi)
- ✅ Statistik pembayaran
- ✅ Multiple payment methods (Cash, Transfer, QRIS)

### Endpoints
```
POST   /api/payments              # Membuat pembayaran baru
GET    /api/payments              # Daftar pembayaran
GET    /api/payments/:id          # Detail pembayaran
PUT    /api/payments/:id/status   # Update status pembayaran
POST   /api/payments/:id/process  # Proses pembayaran
GET    /api/payments/stats        # Statistik pembayaran
```

### Payment Methods
- `cash` - Pembayaran tunai
- `transfer` - Transfer bank
- `qr` - QRIS

### Payment Status
- `pending` - Menunggu pembayaran
- `paid` - Sudah dibayar
- `failed` - Gagal
- `refunded` - Dikembalikan

## 2. Notification System

### Fitur Utama
- ✅ Push notification
- ✅ Email notification (simulasi)
- ✅ Real-time updates
- ✅ Bulk notifications
- ✅ Notification priority levels
- ✅ Read/unread status tracking

### Endpoints
```
POST   /api/notifications              # Membuat notifikasi (admin)
GET    /api/notifications              # Daftar notifikasi user
GET    /api/notifications/:id          # Detail notifikasi
PUT    /api/notifications/:id/read     # Mark as read
PUT    /api/notifications/read-all     # Mark all as read
DELETE /api/notifications/:id          # Hapus notifikasi
GET    /api/notifications/stats        # Statistik notifikasi
POST   /api/admin/notifications/bulk   # Bulk notifications (admin)
```

### Notification Types
- `order` - Notifikasi order
- `payment` - Notifikasi pembayaran
- `system` - Notifikasi sistem
- `promo` - Notifikasi promosi
- `driver` - Notifikasi driver

### Priority Levels
- `low` - Prioritas rendah
- `normal` - Prioritas normal
- `high` - Prioritas tinggi
- `urgent` - Prioritas mendesak

## 3. Location Tracking System

### Fitur Utama
- ✅ Real-time location tracking
- ✅ Driver location updates
- ✅ Nearby drivers search
- ✅ Route calculation
- ✅ Location history
- ✅ Online/offline status

### Endpoints
```
POST   /api/driver/location                    # Update lokasi driver
GET    /api/driver/location                    # Get lokasi driver
PUT    /api/driver/online-status               # Set online status
GET    /api/driver/location/history            # Location history
GET    /api/location/drivers/nearby            # Nearby drivers (public)
GET    /api/location/drivers/:id               # Driver location (public)
GET    /api/location/routes/:order_id          # Driver route (public)
```

### Location Data Structure
```json
{
  "latitude": -7.797068,
  "longitude": 110.370529,
  "accuracy": 10.5,
  "speed": 25.0,
  "heading": 90.0,
  "is_online": true,
  "last_seen": "2024-01-01T12:00:00Z"
}
```

## 4. Enhanced Security & Validation

### Fitur Keamanan
- ✅ Input validation untuk semua endpoint
- ✅ Coordinate validation
- ✅ Payment method validation
- ✅ Notification type validation
- ✅ Role-based access control
- ✅ Rate limiting

### Validation Rules
- Latitude: -90 to 90
- Longitude: -180 to 180
- Payment methods: cash, transfer, qr
- Notification types: order, payment, system, promo, driver
- Priority levels: low, normal, high, urgent

## 5. Database Models

### New Models Added

#### Payment Model
```go
type Payment struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    OrderID   uint           `json:"order_id" gorm:"unique"`
    Amount    float64        `json:"amount" gorm:"not null"`
    Method    PaymentMethod  `json:"method"`
    Status    PaymentStatus  `json:"status"`
    Reference string         `json:"reference"`
    PaidAt    *time.Time     `json:"paid_at"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}
```

#### Notification Model
```go
type Notification struct {
    ID        uint                `json:"id" gorm:"primaryKey"`
    UserID    uint                `json:"user_id" gorm:"not null;index"`
    Title     string              `json:"title" gorm:"not null"`
    Message   string              `json:"message" gorm:"not null"`
    Type      NotificationType    `json:"type"`
    Priority  NotificationPriority `json:"priority"`
    IsRead    bool                `json:"is_read" gorm:"default:false"`
    Data      string              `json:"data" gorm:"type:text"`
    CreatedAt time.Time           `json:"created_at"`
    UpdatedAt time.Time           `json:"updated_at"`
}
```

#### DriverLocation Model
```go
type DriverLocation struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    DriverID  uint           `json:"driver_id" gorm:"not null;uniqueIndex"`
    Latitude  float64        `json:"latitude" gorm:"not null"`
    Longitude float64        `json:"longitude" gorm:"not null"`
    Accuracy  float64        `json:"accuracy"`
    Speed     float64        `json:"speed"`
    Heading   float64        `json:"heading"`
    IsOnline  bool           `json:"is_online" gorm:"default:false"`
    LastSeen  time.Time      `json:"last_seen"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
}
```

## 6. Testing

### Test Coverage
- ✅ Payment creation and validation
- ✅ Notification system testing
- ✅ Location tracking validation
- ✅ Input validation testing
- ✅ Error handling testing
- ✅ Bulk operations testing

### Running Tests
```bash
cd backend
go test ./handlers -v
```

## 7. Integration Examples

### Frontend Integration

#### Payment Processing
```javascript
// Create payment
const createPayment = async (orderId, method, amount) => {
  const response = await fetch('/api/payments', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      order_id: orderId,
      method: method,
      amount: amount
    })
  });
  return response.json();
};
```

#### Real-time Notifications
```javascript
// Get notifications
const getNotifications = async () => {
  const response = await fetch('/api/notifications', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return response.json();
};

// Mark as read
const markAsRead = async (notificationId) => {
  const response = await fetch(`/api/notifications/${notificationId}/read`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return response.json();
};
```

#### Location Tracking
```javascript
// Update driver location
const updateLocation = async (latitude, longitude) => {
  const response = await fetch('/api/driver/location', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      latitude: latitude,
      longitude: longitude,
      accuracy: 10.5,
      speed: 25.0,
      heading: 90.0
    })
  });
  return response.json();
};

// Get nearby drivers
const getNearbyDrivers = async (lat, lng, radius = 5) => {
  const response = await fetch(`/api/location/drivers/nearby?lat=${lat}&lng=${lng}&radius=${radius}`);
  return response.json();
};
```

## 8. Performance Considerations

### Optimization
- ✅ Pagination untuk semua list endpoints
- ✅ Database indexing untuk location queries
- ✅ Efficient coordinate calculations
- ✅ Background processing untuk notifications
- ✅ Caching untuk frequently accessed data

### Monitoring
- ✅ Health check endpoints
- ✅ Metrics collection
- ✅ Error tracking
- ✅ Performance monitoring

## 9. Future Enhancements

### Planned Features
- 🔄 WebSocket integration untuk real-time updates
- 🔄 Push notification service integration
- 🔄 Google Maps API integration
- 🔄 Payment gateway integration
- 🔄 Advanced analytics dashboard
- 🔄 Mobile app API endpoints

### Scalability
- 🔄 Microservices architecture
- 🔄 Load balancing
- 🔄 Database sharding
- 🔄 CDN integration
- 🔄 API versioning

## 10. Deployment

### Environment Variables
```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_NAME=greenbecak
DB_USER=root
DB_PASSWORD=password

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Server
PORT=8080
ENV=development
```

### Docker Deployment
```bash
# Build image
docker build -t greenbecak-backend .

# Run container
docker run -p 8080:8080 greenbecak-backend
```

## 11. API Documentation

### Swagger Documentation
- URL: `http://localhost:8080/swagger/index.html`
- Auto-generated dari code comments
- Interactive testing interface

### Postman Collection
- Available in `/docs/postman/`
- Pre-configured requests
- Environment variables setup

## 12. Support

### Contact
- Email: support@greenbecak.com
- Documentation: `/docs/`
- Issues: GitHub Issues

### Contributing
- Fork repository
- Create feature branch
- Submit pull request
- Follow coding standards

---

**GreenBecak API v2.0** - Enhanced with Payment, Notification, and Location Tracking Systems
