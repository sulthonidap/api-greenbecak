# Coolify Deployment Guide

## Persiapan Deployment

### 1. Environment Variables
Buat file `.env` berdasarkan `coolify.env.example` dan isi dengan nilai yang sesuai:

```bash
cp coolify.env.example .env
```

**Environment Variables yang Wajib Diisi:**
- `DB_HOST`: Host database Anda
- `DB_USER`: Username database
- `DB_PASSWORD`: Password database
- `JWT_SECRET`: Secret key untuk JWT (minimal 16 karakter)

### 2. Database Setup
Pastikan database MySQL/MariaDB sudah tersedia dan dapat diakses dari container.

### 3. Konfigurasi Coolify

#### Di Coolify Dashboard:
1. **Build Pack**: Docker
2. **Dockerfile Path**: `./Dockerfile`
3. **Port**: `8080`
4. **Health Check Path**: `/health`
5. **Environment Variables**: Import dari file `.env`

#### Environment Variables di Coolify:
```
DB_HOST=your-database-host
DB_PORT=3306
DB_USER=your-db-user
DB_PASSWORD=your-db-password
DB_NAME=greenbecak_db
JWT_SECRET=your-super-secret-jwt-key-here
SERVER_PORT=8080
SERVER_MODE=release
CORS_ALLOWED_ORIGINS=*
```

### 4. Troubleshooting

#### Error "no available server":
1. **Pastikan server binding ke 0.0.0.0** ✅ (Sudah diperbaiki)
2. **Health check endpoint berfungsi** ✅ (Sudah diperbaiki)
3. **Environment variables lengkap** ✅ (Sudah diperbaiki)
4. **Port 8080 terbuka** ✅ (Sudah dikonfigurasi)

#### Log yang Perlu Diperhatikan:
```bash
# Cek log container
docker logs greenbecak-api

# Cek health check
curl http://localhost:8080/health
```

#### Common Issues:
1. **Database Connection**: Pastikan database dapat diakses dari container
2. **Port Binding**: Pastikan aplikasi bind ke 0.0.0.0:8080
3. **Health Check**: Pastikan endpoint `/health` mengembalikan status 200
4. **Environment Variables**: Pastikan semua env vars yang diperlukan sudah diset

### 5. Testing Deployment

Setelah deployment berhasil, test endpoint berikut:

```bash
# Health check
curl https://your-domain.com/health

# API documentation
curl https://your-domain.com/swagger/

# Test endpoint
curl https://your-domain.com/api/v1/health
```

### 6. Monitoring

Monitor aplikasi melalui:
- Coolify dashboard logs
- Health check endpoint
- Application metrics (jika tersedia)

## File yang Telah Diperbaiki:

1. **Dockerfile**: 
   - Health check timeout diperpanjang
   - Start period diperpanjang
   - Scripts copy dibuat optional

2. **main.go**:
   - Server binding ke 0.0.0.0 untuk akses eksternal

3. **handlers/health.go**:
   - Health check disederhanakan untuk container orchestration

4. **utils/env.go**:
   - Default values untuk environment variables
   - Validasi yang lebih fleksibel

5. **coolify.yml**:
   - Konfigurasi khusus untuk Coolify deployment

6. **coolify.env.example**:
   - Template environment variables untuk Coolify
