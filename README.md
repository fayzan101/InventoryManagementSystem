<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/PostgreSQL-12+-336791?logo=postgresql" alt="PostgreSQL Version"/>
</p>

# Inventory Management System (IMS)

**A professional, enterprise-grade inventory management backend built with Go, GORM, and PostgreSQL.**

---

## 🚀 Features

- **Product Management** – CRUD operations for products with detailed specs
- **Warehouse Management** – Multi-warehouse, location, and capacity tracking
- **Inventory Tracking** – Real-time stock monitoring, movement history, and low stock alerts
- **Purchase & Sales Orders** – Supplier/customer order management with item-level tracking
- **Audit Logging** – Full audit trail for compliance and accountability
- **Reporting** – Inventory, sales, and stock summary reports
- **RESTful API** – Clean, standard REST endpoints for all operations
- **WebSocket Real-Time Updates** – Live inventory, product, warehouse, and supplier notifications
- **Database Migrations** – Automated schema management with SQL migrations
- **Docker Support** – Run the entire stack with Docker Compose
- **Environment Configuration** – Flexible config via `.env` file

---

## 📦 Prerequisites

- [Go 1.25+](https://golang.org/dl/)
- [PostgreSQL 12+](https://www.postgresql.org/download/)
- [Git](https://git-scm.com/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/) (optional, for containerized setup)

---

## ⚡ Quick Start

### 1. Clone the repository
```bash
git clone <repository-url>
cd Golang
```

### 2. Install Go dependencies
```bash
go mod download
```

### 3. Configure Environment
Create a `.env` file in the project root:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=ims_user
DB_PASSWORD=your_password_here
DB_NAME=ims_db
SERVER_PORT=8080
```

### 4. Set up PostgreSQL database
```bash
createdb ims_db
createuser ims_user -P  # Set a password when prompted
```

### 5. Run Database Migrations
Migrations run automatically on server start. SQL files are in `migrations/`.

### 6. Start the Application
```bash
go run main.go
```
The API will be available at [http://localhost:8080](http://localhost:8080)

---

## 🐳 Docker Compose (Recommended)

Spin up the entire stack (API + PostgreSQL) with one command:

```bash
docker-compose up --build
```

Configure environment variables in your `.env` file. The Go app will be available at [http://localhost:3000](http://localhost:3000) by default.

---

## 🗂️ Project Structure

```
. (root)
├── main.go                  # Application entry point
├── go.mod                   # Go module definition
├── api/                     # API-related code
├── cmd/                     # Command-line interface (Cobra CLI)
├── configs/                 # Configuration management
├── internal/                # Private application code
│   ├── db.go               # Database initialization
│   ├── models.go           # Data models
│   ├── audit/              # Audit logging module
│   ├── inventory/          # Inventory management module
│   ├── orders/             # Sales orders module
│   ├── products/           # Product management module
│   ├── reports/            # Reporting module
│   ├── suppliers/          # Supplier management module
│   └── warehouses/         # Warehouse management module
├── internal/websocket/      # WebSocket real-time backend
├── migrations/              # SQL migration files
├── pkg/                     # Public/reusable packages
├── scripts/                 # Utility scripts
├── test/                    # Test files
├── web/                     # Web assets (HTML clients)
├── docs/                    # Documentation
└── Dockerfile, docker-compose.yml
```

---

## 🔗 API Endpoints

See [docs/API_DOCUMENTATION.md](docs/API_DOCUMENTATION.md) for full details and examples.

Or import [IMS_Postman_Collection.json](IMS_Postman_Collection.json) into Postman for ready-to-use requests.

---

## 🌐 WebSocket Real-Time Integration

- Real-time updates for inventory, products, warehouses, and suppliers
- Connect to endpoints like:
  - `ws://localhost:3000/ws/inventory`
  - `ws://localhost:3000/ws/products`
  - `ws://localhost:3000/ws/warehouses`
  - `ws://localhost:3000/ws/suppliers`
- See [docs/WEBSOCKET_INTEGRATION.md](docs/WEBSOCKET_INTEGRATION.md) and [docs/WEBSOCKET_DOCUMENTATION.md](docs/WEBSOCKET_DOCUMENTATION.md) for message formats and integration guide
- Example HTML clients in `web/` folder (e.g., `inventory-realtime.html`)

---

## 🧪 Testing

Run all tests:
```bash
go test ./...
```
Or use the provided script:
```bash
bash scripts/test.sh
```

---

## 📝 Documentation

- [API Documentation](docs/API_DOCUMENTATION.md)
- [WebSocket Integration Guide](docs/WEBSOCKET_INTEGRATION.md)
- [WebSocket Protocol Details](docs/WEBSOCKET_DOCUMENTATION.md)

---

## 🛠️ Development & Contribution

1. Fork and clone the repository
2. Create a new branch for your feature or bugfix
3. Follow Go best practices (package organization, error handling, interfaces)
4. Add tests for new features
5. Update documentation as needed
6. Submit a pull request

---

## 🔒 Security & Production

- Restrict WebSocket origins in production (see `internal/websocket/handler.go`)
- Use `wss://` and SSL/TLS for secure connections
- Add authentication and rate limiting as needed

---

For questions or support, see the documentation or contact the development team.
