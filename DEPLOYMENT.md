# GreenBecak Backend Deployment Guide

## Overview

Guide ini menjelaskan cara deploy GreenBecak Backend ke berbagai environment.

## Prerequisites

- Go 1.21+
- MySQL 8.0+
- Docker (optional)
- Git

## Local Development

### 1. Setup Environment

```bash
# Clone repository
git clone <repository-url>
cd greenbecak/backend

# Install dependencies
go mod tidy

# Copy environment file
cp env.example .env

# Edit environment variables
nano .env
```

### 2. Database Setup

```sql
-- Create database
CREATE DATABASE greenbecak_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create user (optional)
CREATE USER 'greenbecak'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON greenbecak_db.* TO 'greenbecak'@'localhost';
FLUSH PRIVILEGES;
```

### 3. Run Application

```bash
# Development mode
go run main.go

# Or using Makefile
make run
```

### 4. Seed Data

```bash
# Run seeding script
go run scripts/seed.go

# Or using Makefile
make seed
```

## Docker Deployment

### 1. Build Docker Image

```bash
# Build image
docker build -t greenbecak-backend .

# Or using Makefile
make docker-build
```

### 2. Run with Docker

```bash
# Run container
docker run -p 8080:8080 --env-file .env greenbecak-backend

# Or using Makefile
make docker-run
```

### 3. Docker Compose

```bash
# Start all services
docker-compose up -d

# Or using Makefile
make docker-compose-up

# Stop services
docker-compose down

# Or using Makefile
make docker-compose-down
```

## Production Deployment

### 1. Environment Variables

Create `.env` file for production:

```env
# Database Configuration
DB_HOST=your-db-host
DB_PORT=3306
DB_USER=your-db-user
DB_PASSWORD=your-secure-password
DB_NAME=greenbecak_db

# JWT Secret (generate secure secret)
JWT_SECRET=your-very-secure-jwt-secret-at-least-32-characters

# Server Configuration
SERVER_PORT=8080
SERVER_MODE=release

# CORS Configuration
CORS_ALLOWED_ORIGINS=https://your-frontend-domain.com
```

### 2. Build for Production

```bash
# Build for Linux
make build-linux

# Build for Windows
make build-windows

# Build for macOS
make build-macos
```

### 3. Systemd Service (Linux)

Create service file `/etc/systemd/system/greenbecak-backend.service`:

```ini
[Unit]
Description=GreenBecak Backend API
After=network.target

[Service]
Type=simple
User=greenbecak
WorkingDirectory=/opt/greenbecak-backend
ExecStart=/opt/greenbecak-backend/greenbecak-backend
Restart=always
RestartSec=5
EnvironmentFile=/opt/greenbecak-backend/.env

[Install]
WantedBy=multi-user.target
```

Enable and start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable greenbecak-backend
sudo systemctl start greenbecak-backend
sudo systemctl status greenbecak-backend
```

### 4. Nginx Reverse Proxy

Create nginx configuration `/etc/nginx/sites-available/greenbecak-backend`:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable site:

```bash
sudo ln -s /etc/nginx/sites-available/greenbecak-backend /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 5. SSL with Let's Encrypt

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

## Kubernetes Deployment

### 1. Create Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: greenbecak
```

### 2. ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: greenbecak-config
  namespace: greenbecak
data:
  DB_HOST: "mysql-service"
  DB_PORT: "3306"
  DB_NAME: "greenbecak_db"
  SERVER_PORT: "8080"
  SERVER_MODE: "release"
```

### 3. Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: greenbecak-secret
  namespace: greenbecak
type: Opaque
data:
  DB_USER: <base64-encoded-username>
  DB_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-jwt-secret>
```

### 4. Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: greenbecak-backend
  namespace: greenbecak
spec:
  replicas: 3
  selector:
    matchLabels:
      app: greenbecak-backend
  template:
    metadata:
      labels:
        app: greenbecak-backend
    spec:
      containers:
      - name: greenbecak-backend
        image: greenbecak-backend:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: greenbecak-config
        - secretRef:
            name: greenbecak-secret
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### 5. Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: greenbecak-backend-service
  namespace: greenbecak
spec:
  selector:
    app: greenbecak-backend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

### 6. Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: greenbecak-backend-ingress
  namespace: greenbecak
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: greenbecak-backend-service
            port:
              number: 80
```

## Monitoring & Logging

### 1. Health Checks

```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Liveness check
curl http://localhost:8080/live
```

### 2. Metrics

```bash
# Get metrics
curl http://localhost:8080/metrics

# Reset metrics
curl -X POST http://localhost:8080/metrics/reset
```

### 3. Alerts

```bash
# Get all alerts
curl http://localhost:8080/alerts

# Get active alerts
curl http://localhost:8080/alerts/active
```

### 4. Logs

```bash
# View logs
tail -f logs/app.log

# Systemd logs
sudo journalctl -u greenbecak-backend -f

# Docker logs
docker logs -f greenbecak-backend
```

## Database Management

### 1. Backup

```bash
# Create backup
make backup

# List backups
make backup-list

# Cleanup old backups
make backup-cleanup
```

### 2. Restore

```bash
# Restore from backup
make backup-restore file=backups/greenbecak_backup_20240101_120000.sql
```

### 3. Migration

```bash
# Run migrations
make migrate

# Seed data
make seed
```

## Security

### 1. Firewall

```bash
# Allow only necessary ports
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

### 2. SSL/TLS

```bash
# Generate SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### 3. Database Security

```sql
-- Create dedicated user
CREATE USER 'greenbecak'@'localhost' IDENTIFIED BY 'strong-password';
GRANT SELECT, INSERT, UPDATE, DELETE ON greenbecak_db.* TO 'greenbecak'@'localhost';
FLUSH PRIVILEGES;
```

## Performance Tuning

### 1. Database Optimization

```sql
-- Add indexes for better performance
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created_at);
CREATE INDEX idx_drivers_status ON drivers(status);
```

### 2. Application Tuning

```env
# Increase connection pool
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=3600

# Enable compression
ENABLE_GZIP=true
```

### 3. Nginx Optimization

```nginx
# Enable gzip compression
gzip on;
gzip_types text/plain text/css application/json application/javascript;

# Enable caching
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

## Troubleshooting

### 1. Common Issues

**Database Connection Failed**
```bash
# Check database status
sudo systemctl status mysql

# Check connection
mysql -u greenbecak -p greenbecak_db
```

**Application Won't Start**
```bash
# Check logs
tail -f logs/app.log

# Check environment variables
cat .env

# Test configuration
go run main.go --config-test
```

**High Memory Usage**
```bash
# Check memory usage
free -h

# Check application memory
ps aux | grep greenbecak-backend
```

### 2. Debug Mode

```bash
# Enable debug mode
export SERVER_MODE=debug

# Run with verbose logging
go run main.go -v
```

### 3. Performance Monitoring

```bash
# Monitor CPU and memory
htop

# Monitor network
iftop

# Monitor disk I/O
iotop
```

## Backup & Recovery

### 1. Automated Backups

Create backup script `/opt/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="/opt/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/greenbecak_backup_$DATE.sql"

# Create backup
mysqldump -u greenbecak -p'password' greenbecak_db > $BACKUP_FILE

# Compress backup
gzip $BACKUP_FILE

# Cleanup old backups (keep 7 days)
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_FILE.gz"
```

Make executable and add to cron:

```bash
chmod +x /opt/backup.sh
crontab -e
# Add: 0 2 * * * /opt/backup.sh
```

### 2. Disaster Recovery

```bash
# Stop application
sudo systemctl stop greenbecak-backend

# Restore database
mysql -u greenbecak -p greenbecak_db < backup_file.sql

# Start application
sudo systemctl start greenbecak-backend
```

## Scaling

### 1. Horizontal Scaling

```bash
# Scale deployment
kubectl scale deployment greenbecak-backend --replicas=5

# Or with Docker Compose
docker-compose up --scale api=3
```

### 2. Load Balancer

```yaml
apiVersion: v1
kind: Service
metadata:
  name: greenbecak-backend-service
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: greenbecak-backend
```

### 3. Auto Scaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: greenbecak-backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: greenbecak-backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Support

Untuk bantuan deployment:

- **Documentation**: https://docs.greenbecak.com/deployment
- **GitHub Issues**: https://github.com/greenbecak/backend/issues
- **Email Support**: support@greenbecak.com
- **Slack**: #greenbecak-support
