# API Gateway API Documentation

This document provides detailed information about all available endpoints in the API Gateway.

## Base URL

- **Development**: `http://localhost:8080`
- **Production**: Configure according to your deployment

## Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

## Endpoints

### Health Check

#### GET /health

Check if the gateway is running.

**Response:**
```json
{
  "status": "healthy"
}
```

### Authentication

#### POST /login

Authenticate a user and receive a JWT token.

**Request Body:**
```json
{
  "username": "your_username",
  "password": "your_password"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "userID": "user_uuid",
  "username": "your_username"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Invalid credentials

### Reverse Proxy

#### ANY /proxy/{service_name}/*proxyPath

Forward requests to backend services. The gateway will route the request to the appropriate backend service based on the service name.

**Headers Required:**
- `Authorization: Bearer <jwt_token>`

**Example:**
```
GET /proxy/user-service/api/users
Authorization: Bearer <jwt_token>
```

**Response:** The response from the backend service.

**Error Responses:**
- `401 Unauthorized`: Missing or invalid JWT token
- `429 Too Many Requests`: Rate limit exceeded
- `502 Bad Gateway`: Backend service unavailable
- `404 Not Found`: Service not found

### Admin Endpoints

All admin endpoints require JWT authentication with admin privileges.

#### GET /admin/routes

Get all configured routes.

**Response:**
```json
[
  {
    "id": 1,
    "service_name": "user-service",
    "backend_url": "http://user-service:8080",
    "rate_limit": 100,
    "rate_limit_window": 60,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### POST /admin/routes

Create a new route.

**Request Body:**
```json
{
  "service_name": "user-service",
  "backend_url": "http://user-service:8080",
  "rate_limit": 100,
  "rate_limit_window": 60
}
```

**Response:**
```json
{
  "id": 1,
  "service_name": "user-service",
  "backend_url": "http://user-service:8080",
  "rate_limit": 100,
  "rate_limit_window": 60,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### PUT /admin/routes/{id}

Update an existing route.

**Request Body:**
```json
{
  "service_name": "user-service",
  "backend_url": "http://user-service:8080",
  "rate_limit": 150,
  "rate_limit_window": 60
}
```

**Response:** Updated route object

#### DELETE /admin/routes/{id}

Delete a route.

**Response:**
```json
{
  "message": "Route deleted successfully"
}
```

#### GET /admin/routes/stats

Get route statistics.

**Response:**
```json
{
  "total_routes": 5,
  "active_routes": 4,
  "total_requests": 1000,
  "average_response_time": 150
}
```

#### GET /admin/routes/optimizer/stats

Get route optimizer statistics.

**Response:**
```json
{
  "hash_map_size": 5,
  "prefix_tree_size": 12,
  "last_updated": "2024-01-01T00:00:00Z"
}
```

#### POST /admin/routes/optimizer/benchmark

Benchmark route lookup performance.

**Request Body:**
```json
{
  "service_names": ["user-service", "order-service", "payment-service"]
}
```

**Response:**
```json
{
  "hash_map_microseconds": 50,
  "prefix_tree_microseconds": 150,
  "optimized_microseconds": 75,
  "improvement_percentage": 25.0
}
```

## Rate Limiting

The gateway implements rate limiting using a token bucket algorithm with Redis. Each route can have its own rate limit configuration:

- **rate_limit**: Maximum number of requests allowed
- **rate_limit_window**: Time window in seconds

When rate limit is exceeded, the gateway returns:
- **Status Code**: `429 Too Many Requests`
- **Headers**: 
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time until reset

## Error Handling

All endpoints return consistent error responses:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `200 OK`: Success
- `201 Created`: Resource created
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Examples

### Complete Workflow

1. **Login to get token:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'
```

2. **Create a route:**
```bash
curl -X POST http://localhost:8080/admin/routes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "user-service",
    "backend_url": "http://user-service:8080",
    "rate_limit": 100,
    "rate_limit_window": 60
  }'
```

3. **Make a request through the gateway:**
```bash
curl -X GET http://localhost:8080/proxy/user-service/api/users \
  -H "Authorization: Bearer <token>"
```

4. **Check route statistics:**
```bash
curl -X GET http://localhost:8080/admin/routes/stats \
  -H "Authorization: Bearer <token>"
``` 