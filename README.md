# API Gateway with Rate Limiting & JWT Validation

A robust, production-ready API Gateway built with Go that provides reverse proxy routing, JWT validation, and rate limiting capabilities. This gateway acts as a single entry point for your microservices architecture, offering security, monitoring, and traffic control features.

## 🚀 Features

- **Reverse Proxy Routing**: Dynamic routing to backend services with optimized lookup
- **JWT Authentication**: Secure token-based authentication with middleware
- **Rate Limiting**: Redis-based token bucket algorithm with per-route configuration
- **Dynamic Route Management**: Admin APIs to manage routes on-the-fly
- **Route Optimization**: Hash-map and prefix-tree structures for 20% faster lookups
- **Comprehensive Testing**: httptest-based test suite covering critical paths
- **Monitoring & Logging**: Request/response logging and performance metrics
- **Docker Support**: Containerized deployment with Docker Compose
- **Database Integration**: PostgreSQL for route metadata storage

## 🛠 Tech Stack

- **Backend**: Go (Gin Framework)
- **Database**: PostgreSQL with GORM
- **Caching**: Redis
- **Containerization**: Docker & Docker Compose
- **Authentication**: JWT (JSON Web Tokens)
- **Testing**: Go's net/http/httptest

## 📁 Project Structure

```
.
├── cmd/
│   └── gateway/
│       └── main.go                 # Main gateway application
├── internal/
│   ├── config/
│   │   ├── config.go              # Configuration management
│   │   └── redis.go               # Redis client setup
│   ├── middleware/
│   │   ├── auth/
│   │   │   └── jwt.go             # JWT authentication middleware
│   │   └── ratelimit/
│   │       └── ratelimit.go       # Rate limiting middleware
│   ├── models/
│   │   └── user.go                # User model
│   └── services/
│       ├── admin_api.go           # Admin API endpoints
│       ├── reverse_proxy.go       # Reverse proxy handler
│       └── route_optimizer.go     # Route optimization engine
├── tests/
│   └── gateway_test.go            # Comprehensive test suite
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── configs/
│   └── config.yaml
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Redis 7.0 or higher
- PostgreSQL 14 or higher

### Running with Docker Compose

1. Clone the repository:
   ```bash
   git clone https://github.com/kart2405/API_Gateway.git
   cd api-gateway
   ```

2. Create a `.env` file:
   ```bash
   cp .env.example .env
   ```

3. Start the services:
   ```bash
   docker-compose up -d
   ```

4. The API Gateway will be available at `http://localhost:8080`

### Running Tests

```bash
# Run all tests
go test ./tests/...

# Run tests with verbose output
go test -v ./tests/...

# Run benchmarks
go test -bench=. ./tests/...

# Run specific test
go test -run TestLoginEndpoint ./tests/
```

## 🔄 How It Works

1. **Request Flow**:
   - Client sends request to the gateway
   - Gateway validates JWT token
   - Rate limiting check is performed (per-route configuration)
   - Request is routed using optimized lookup (hash-map + prefix-tree)
   - Response is returned to client

2. **Rate Limiting**:
   - Uses Redis-based token bucket algorithm
   - Configurable limits per route/service
   - Distributed rate limiting support
   - Per-user and per-service rate limiting

3. **Route Optimization**:
   - Hash-map for O(1) exact matches
   - Prefix-tree for pattern matching
   - 20% performance improvement over simple map lookup
   - Automatic fallback between optimization strategies

4. **Admin APIs**:
   - CRUD operations for route management
   - Real-time route statistics
   - Performance benchmarking
   - Dynamic route updates without restart

## 📝 API Endpoints

### Public Endpoints

#### Login
```bash
POST /login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

### Protected Endpoints

#### Reverse Proxy
```bash
GET /proxy/{service_name}/{path}
Authorization: Bearer {jwt_token}
```

### Admin Endpoints (JWT Required)

#### Get All Routes
```bash
GET /admin/routes
Authorization: Bearer {admin_token}
```

#### Create Route
```bash
POST /admin/routes
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "service_name": "user-service",
  "backend_url": "http://user-service:8080",
  "rate_limit": 100,
  "rate_limit_window": 60
}
```

#### Update Route
```bash
PUT /admin/routes/{id}
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "rate_limit": 150
}
```

#### Delete Route
```bash
DELETE /admin/routes/{id}
Authorization: Bearer {admin_token}
```

#### Get Route Statistics
```bash
GET /admin/routes/stats
Authorization: Bearer {admin_token}
```

#### Get Optimizer Statistics
```bash
GET /admin/routes/optimizer/stats
Authorization: Bearer {admin_token}
```

#### Benchmark Route Lookup
```bash
POST /admin/routes/optimizer/benchmark
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "service_names": ["user-service", "order-service", "payment-service"]
}
```

## 🧪 Testing

The project includes comprehensive tests using Go's `httptest` package:

- **Authentication Tests**: JWT token validation and generation
- **Rate Limiting Tests**: Token bucket algorithm verification
- **Admin API Tests**: CRUD operations for route management
- **Route Optimizer Tests**: Hash-map and prefix-tree functionality
- **Reverse Proxy Tests**: Request forwarding and error handling
- **Performance Benchmarks**: Route lookup optimization measurements

### Running Tests

```bash
# Run all tests
go test ./tests/...

# Run with coverage
go test -cover ./tests/...

# Run benchmarks
go test -bench=. ./tests/
```

## 📊 Performance Optimization

### Route Lookup Optimization

The gateway implements a dual-strategy approach for route matching:

1. **Hash-Map Lookup**: O(1) time complexity for exact matches
2. **Prefix-Tree Lookup**: Efficient pattern matching for complex routes
3. **Smart Fallback**: Automatic selection of the most efficient method

### Benchmark Results

Typical performance improvements:
- **Hash-Map**: ~0.1μs per lookup
- **Prefix-Tree**: ~0.3μs per lookup  
- **Optimized**: ~0.15μs per lookup (with fallback)
- **Improvement**: ~20% faster than simple map lookup

## 🗺 Roadmap

- [x] Basic reverse proxy functionality
- [x] JWT validation middleware
- [x] Redis-based rate limiting
- [x] Admin API for route management
- [x] Docker support
- [x] Route optimization (hash-map + prefix-tree)
- [x] Comprehensive testing with httptest
- [x] Performance benchmarking
- [ ] Circuit breaker implementation
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Metrics dashboard
- [ ] Load balancing
- [ ] WebSocket support
- [ ] GraphQL support
- [ ] Caching layer
- [ ] OAuth2 integration

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Built with ❤️ using Go and Gin
