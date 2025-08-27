# CORS Setup Guide

## Masalah CORS di Production

Jika API Anda berjalan dengan baik di localhost tapi mengalami error CORS ketika diakses dari domain production dengan SSL, ini adalah panduan untuk mengatasinya.

## Konfigurasi CORS

### 1. Environment Variable

Tambahkan atau update environment variable `CORS_ALLOWED_ORIGINS` di file `.env`:

```env
# Development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Production (ganti dengan domain Anda)
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com,https://app.yourdomain.com
```

### 2. Format Domain

Pastikan format domain sesuai dengan protokol yang digunakan:

- **HTTP**: `http://localhost:3000`
- **HTTPS**: `https://yourdomain.com`
- **Multiple domains**: Pisahkan dengan koma tanpa spasi

### 3. Contoh Konfigurasi

```env
# Untuk development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000

# Untuk production
CORS_ALLOWED_ORIGINS=https://greenbecak.com,https://www.greenbecak.com,https://app.greenbecak.com

# Untuk testing dengan subdomain
CORS_ALLOWED_ORIGINS=https://staging.greenbecak.com,https://dev.greenbecak.com
```

## Deployment

### Docker Compose

Update `docker-compose.yml`:

```yaml
environment:
  - CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### Production Server

Set environment variable di server production:

```bash
export CORS_ALLOWED_ORIGINS="https://yourdomain.com,https://www.yourdomain.com"
```

### Systemd Service

Update file service `/etc/systemd/system/greenbecak-backend.service`:

```ini
[Service]
Environment="CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com"
```

## Troubleshooting

### 1. Check CORS Configuration

API akan menampilkan log CORS configuration saat startup:

```
CORS Allowed Origins: [https://yourdomain.com https://www.yourdomain.com]
```

### 2. Browser Developer Tools

Buka Developer Tools (F12) dan lihat tab Console untuk error CORS:

```
Access to fetch at 'https://api.yourdomain.com/api/auth/login' from origin 'https://yourdomain.com' has been blocked by CORS policy
```

### 3. Test dengan Postman

Jika request berhasil di Postman tapi gagal di browser, ini menandakan masalah CORS.

### 4. Common Issues

#### Domain tidak terdaftar
```
Error: Origin 'https://yourdomain.com' is not allowed
```
**Solusi**: Tambahkan domain ke `CORS_ALLOWED_ORIGINS`

#### Protocol mismatch
```
Error: Mixed content - HTTP vs HTTPS
```
**Solusi**: Pastikan semua domain menggunakan protokol yang sama (HTTP atau HTTPS)

#### Subdomain tidak terdaftar
```
Error: Origin 'https://app.yourdomain.com' is not allowed
```
**Solusi**: Tambahkan subdomain ke `CORS_ALLOWED_ORIGINS`

## Security Considerations

### 1. Jangan Gunakan Wildcard

❌ **Jangan gunakan**:
```env
CORS_ALLOWED_ORIGINS=*
```

✅ **Gunakan domain spesifik**:
```env
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### 2. Environment-specific Configuration

```env
# Development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173

# Staging
CORS_ALLOWED_ORIGINS=https://staging.yourdomain.com

# Production
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### 3. Regular Review

- Review `CORS_ALLOWED_ORIGINS` secara berkala
- Hapus domain yang tidak digunakan
- Update ketika ada domain baru

## Testing

### 1. Local Testing

```bash
# Test dengan curl
curl -H "Origin: https://yourdomain.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     https://api.yourdomain.com/api/auth/login
```

### 2. Browser Testing

```javascript
// Test di browser console
fetch('https://api.yourdomain.com/api/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'test',
    password: 'test'
  })
})
.then(response => response.json())
.then(data => console.log(data))
.catch(error => console.error('CORS Error:', error));
```

## Monitoring

### 1. Log Monitoring

Monitor log untuk CORS errors:

```bash
tail -f logs/app.log | grep -i cors
```

### 2. Metrics

API menyediakan metrics untuk CORS requests:

```
GET /metrics
```

## Support

Jika masih mengalami masalah CORS:

1. Check log server untuk error details
2. Verify environment variable sudah benar
3. Restart server setelah update konfigurasi
4. Clear browser cache
5. Test dengan incognito mode

---

**Note**: CORS adalah security feature browser, bukan masalah server. Pastikan konfigurasi sesuai dengan security policy aplikasi Anda.
