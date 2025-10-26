# 🚀 Deployment Guide

This guide explains how to build and deploy the Crypto Quant application with embedded frontend.

## 📋 Build Methods

### Method 1: Go Embed (Recommended ⭐)

The frontend is embedded directly into the Go binary using Go's `embed` package. This creates a single binary that contains both backend and frontend.

**Advantages:**
- ✅ Single binary deployment
- ✅ No external file dependencies
- ✅ Easy to distribute
- ✅ Consistent deployment

**Build Commands:**

```bash
# From backend directory
cd backend
make build-full

# Or use the build script from project root
cd ..
./scripts/build.sh
```

**What happens:**
1. Frontend is built using `pnpm build` → creates `frontend/build/`
2. Built files are copied to `backend/internal/api/frontend/build/`
3. Go embeds these files into the binary during compilation
4. Single binary is created: `backend/bin/api`

**Run:**
```bash
cd backend
./bin/api

# Access at http://localhost:8080
# - Frontend: http://localhost:8080
# - API: http://localhost:8080/api/v1
# - Swagger: http://localhost:8080/swagger/index.html
```

---

### Method 2: Docker (Coming Soon)

For containerized deployment with Docker.

---

### Method 3: Separate Frontend Hosting (Coming Soon)

For deploying frontend and backend separately (e.g., frontend on CDN, backend on server).

---

## 🔧 Development Mode

During development, run frontend and backend separately:

```bash
# Terminal 1: Backend
cd backend
make dev-api

# Terminal 2: Frontend
cd frontend
pnpm dev
```

In this mode:
- Frontend runs on http://localhost:5173 with hot reload
- Backend runs on http://localhost:8080
- CORS is enabled for cross-origin requests

---

## 📦 Build Artifacts

After `make build-full`, you'll have:

```
backend/
├── bin/
│   ├── api          # API server with embedded frontend
│   ├── collector    # Data collector
│   └── backtest     # Backtest engine
└── internal/api/frontend/build/  # Embedded frontend files
```

---

## 🛠️ Makefile Commands

From `backend/` directory:

| Command | Description |
|---------|-------------|
| `make build-full` | Build frontend + all binaries (production) |
| `make build-frontend` | Build frontend only |
| `make build` | Build backend binaries only |
| `make build-api` | Build API server only |
| `make clean-all` | Remove all build artifacts |
| `make clean-frontend` | Remove frontend build |
| `make help` | Show all available commands |

---

## 📝 How Go Embed Works

### 1. Frontend Build
```bash
cd frontend
pnpm build
# Creates: frontend/build/
#   ├── index.html
#   ├── _app/
#   │   ├── immutable/
#   │   └── version.json
#   └── ...
```

### 2. Copy to Backend
```bash
cp -r frontend/build/* backend/internal/api/frontend/build/
```

### 3. Go Embed Directive
In `backend/internal/api/embed.go`:
```go
//go:embed frontend/build/*
var frontendFS embed.FS
```

This directive tells Go to embed all files from `frontend/build/` directory into the binary at compile time.

### 4. Serving Static Files
In `backend/internal/api/api.go`:
- Static files (CSS, JS, images) are served directly from embedded FS
- SPA routing: All non-API routes serve `index.html` for client-side routing
- API routes (`/api/v1/*`) are not affected

---

## 🌐 Production Deployment

### Option 1: Direct Binary Deployment

1. Build the binary on your build server:
```bash
./scripts/build.sh
```

2. Copy the binary to your production server:
```bash
scp backend/bin/api user@server:/opt/crypto-quant/
```

3. Run on production server:
```bash
cd /opt/crypto-quant
./api --port 8080 --db /data/trading.db
```

### Option 2: Systemd Service

Create `/etc/systemd/system/crypto-quant.service`:

```ini
[Unit]
Description=Crypto Quant API Server
After=network.target

[Service]
Type=simple
User=cryptoquant
WorkingDirectory=/opt/crypto-quant
ExecStart=/opt/crypto-quant/api --port 8080 --db /data/trading.db
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable crypto-quant
sudo systemctl start crypto-quant
sudo systemctl status crypto-quant
```

---

## 🔍 Troubleshooting

### Frontend files not found

If you see "Frontend not available" error:

1. Check if frontend was built:
```bash
ls -la backend/internal/api/frontend/build/
```

2. Rebuild frontend:
```bash
cd backend
make clean-frontend build-frontend
```

3. Rebuild API binary:
```bash
make build-api
```

### Go embed errors

If you get embed errors during build:

```
pattern frontend/build/*: no matching files found
```

Solution: Build frontend first!
```bash
make build-frontend
```

---

## 📊 Binary Size

The embedded binary will be larger due to frontend assets:

- Backend only: ~20-30 MB
- Backend + Frontend: ~25-35 MB

The size increase is minimal and acceptable for single-binary deployment benefits.

---

## 🎯 Best Practices

1. **Always run `make build-full`** for production builds
2. **Test the binary locally** before deploying to production
3. **Use version tags** for releases
4. **Keep frontend builds separate** in version control (add to .gitignore)
5. **Monitor binary size** if adding large assets to frontend

---

## 📌 Important Notes

- The `internal/api/frontend/build/` directory should be in `.gitignore`
- Frontend must be rebuilt whenever you change frontend code
- Backend must be rebuilt to include new frontend changes
- During development, use separate dev servers for hot reload
- In production, use the single embedded binary

---

## 🔗 Related Documentation

- [API Documentation](./API_BACKTEST.md)
- [Backtest Guide](./BACKTEST.md)
- [Usage Guide](./USAGE_GUIDE.md)






