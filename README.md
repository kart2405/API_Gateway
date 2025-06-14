# API Gateway with Rate Limiting & JWT Validation

A robust, production-ready API Gateway built with Go that provides reverse proxy routing, JWT validation, and rate limiting capabilities. This gateway acts as a single entry point for your microservices architecture, offering security, monitoring, and traffic control features.

## 🚀 Features

- **Reverse Proxy Routing**: Dynamic routing to backend services
- **JWT Authentication**: Secure token-based authentication
- **Rate Limiting**: Redis-based token bucket algorithm for API protection
- **Dynamic Route Management**: Admin APIs to manage routes on-the-fly
- **Monitoring & Logging**: Request/response logging and basic metrics
- **Docker Support**: Containerized deployment with Docker Compose
- **Database Integration**: PostgreSQL for route metadata storage

## 🛠 Tech Stack~

- **Backend**: Go (Gin Framework)
- **Database**: PostgreSQL
- **Caching**: Redis
- **Containerization**: Docker & Docker Compose
- **Authentication**: JWT (JSON Web Tokens)

## 📁 Project Structure

```
.
├── cmd/
│   └── gateway/
│       └── main.go
├── internal/
│   ├── config/
│   ├── middleware/
│   │   ├── auth/
│   │   └── ratelimit/
│   ├── models/
│   ├── routes/
│   └── services/
├── pkg/
│   ├── logger/
│   └── utils/
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

## 🔄 How It Works

1. **Request Flow**:
   - Client sends request to the gateway
   - Gateway validates JWT token
   - Rate limiting check is performed
   - Request is routed to appropriate backend service
   - Response is returned to client

2. **Rate Limiting**:
   - Uses Redis-based token bucket algorithm
   - Configurable limits per route/client
   - Distributed rate limiting support

3. **Route Management**:
   - Routes stored in PostgreSQL
   - Dynamic route updates without restart
   - Support for path-based and host-based routing

## 📝 Admin API Examples

### Add New Route

```bash
curl -X POST http://localhost:8080/admin/routes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "path": "/api/v1/users",
    "target": "http://user-service:8080",
    "methods": ["GET", "POST"],
    "rate_limit": {
      "requests_per_minute": 100
    },
    "auth_required": true
  }'
```

## 🗺 Roadmap

- [x] Basic reverse proxy functionality
- [x] JWT validation middleware
- [x] Redis-based rate limiting
- [x] Admin API for route management
- [x] Docker support
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
# API_Gateway
# API_Gateway
