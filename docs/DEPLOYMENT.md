# Deployment Guide

This guide covers deploying the API Gateway in different environments.

## Prerequisites

- Docker and Docker Compose
- PostgreSQL 14+
- Redis 7+
- Go 1.24.2+ (for local development)

## Environment Variables

The following environment variables can be configured:

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | PostgreSQL username |
| `DB_PASSWORD` | 2405 | PostgreSQL password |
| `DB_NAME` | apigateway | PostgreSQL database name |
| `REDIS_ADDR` | localhost:6379 | Redis address |
| `JWT_SECRET` | your-secret-key | JWT signing secret |
| `GATEWAY_PORT` | 8080 | Gateway port |

## Docker Deployment

### Using Docker Compose (Recommended)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/kart2405/API_Gateway.git
   cd API_gateway_with_ratelimiting_and_Jwt_validation
   ```

2. **Create environment file:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start services:**
   ```bash
   cd docker
   docker-compose up -d
   ```

4. **Verify deployment:**
   ```bash
   curl http://localhost:8080/health
   ```

### Using Docker Only

1. **Build the image:**
   ```bash
   docker build -f docker/Dockerfile -t api-gateway .
   ```

2. **Run the container:**
   ```bash
   docker run -d \
     --name api-gateway \
     -p 8080:8080 \
     -e DB_HOST=your-db-host \
     -e DB_PASSWORD=your-db-password \
     -e REDIS_ADDR=your-redis-host:6379 \
     api-gateway
   ```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster
- kubectl configured
- Helm (optional)

### Using kubectl

1. **Create namespace:**
   ```bash
   kubectl create namespace api-gateway
   ```

2. **Create ConfigMap:**
   ```yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: api-gateway-config
     namespace: api-gateway
   data:
     config.yaml: |
       routes:
         service1: http://service1:8080
         service2: http://service2:8080
   ```

3. **Create Secret:**
   ```bash
   kubectl create secret generic api-gateway-secret \
     --from-literal=db-password=your-password \
     --from-literal=jwt-secret=your-jwt-secret \
     -n api-gateway
   ```

4. **Deploy the application:**
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: api-gateway
     namespace: api-gateway
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: api-gateway
     template:
       metadata:
         labels:
           app: api-gateway
       spec:
         containers:
         - name: api-gateway
           image: api-gateway:latest
           ports:
           - containerPort: 8080
           env:
           - name: DB_HOST
             value: "postgres-service"
           - name: DB_PASSWORD
             valueFrom:
               secretKeyRef:
                 name: api-gateway-secret
                 key: db-password
           - name: REDIS_ADDR
             value: "redis-service:6379"
           livenessProbe:
             httpGet:
               path: /health
               port: 8080
             initialDelaySeconds: 30
             periodSeconds: 10
           readinessProbe:
             httpGet:
               path: /health
               port: 8080
             initialDelaySeconds: 5
             periodSeconds: 5
   ```

5. **Create Service:**
   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: api-gateway-service
     namespace: api-gateway
   spec:
     selector:
       app: api-gateway
     ports:
     - protocol: TCP
       port: 80
       targetPort: 8080
     type: LoadBalancer
   ```

### Using Helm

1. **Create values.yaml:**
   ```yaml
   replicaCount: 3
   
   image:
     repository: api-gateway
     tag: latest
     pullPolicy: IfNotPresent
   
   service:
     type: LoadBalancer
     port: 80
     targetPort: 8080
   
   env:
     DB_HOST: postgres-service
     DB_PORT: 5432
     DB_USER: postgres
     DB_NAME: apigateway
     REDIS_ADDR: redis-service:6379
   
   resources:
     limits:
       cpu: 500m
       memory: 512Mi
     requests:
       cpu: 250m
       memory: 256Mi
   ```

2. **Deploy with Helm:**
   ```bash
   helm install api-gateway ./helm-chart -f values.yaml
   ```

## Production Considerations

### Security

1. **Use strong JWT secrets:**
   ```bash
   openssl rand -base64 32
   ```

2. **Enable HTTPS:**
   - Use a reverse proxy (nginx, traefik)
   - Configure SSL certificates
   - Enable HSTS headers

3. **Network security:**
   - Use private networks for database connections
   - Implement firewall rules
   - Use VPN for admin access

### Performance

1. **Resource limits:**
   - Set appropriate CPU and memory limits
   - Monitor resource usage
   - Scale horizontally as needed

2. **Caching:**
   - Configure Redis for optimal performance
   - Use connection pooling
   - Implement response caching

3. **Monitoring:**
   - Set up health checks
   - Monitor response times
   - Track error rates

### High Availability

1. **Database:**
   - Use managed PostgreSQL service
   - Set up read replicas
   - Implement backup strategies

2. **Redis:**
   - Use Redis Cluster or Sentinel
   - Configure persistence
   - Monitor memory usage

3. **Load Balancing:**
   - Use multiple gateway instances
   - Implement sticky sessions if needed
   - Configure health checks

## Monitoring and Logging

### Health Checks

The gateway provides a health endpoint:
```bash
curl http://your-gateway:8080/health
```

### Metrics

Consider implementing:
- Prometheus metrics
- Custom business metrics
- Performance dashboards

### Logging

Configure structured logging:
- JSON format for production
- Log levels (DEBUG, INFO, WARN, ERROR)
- Log aggregation (ELK stack, Fluentd)

## Troubleshooting

### Common Issues

1. **Database connection failed:**
   - Check database credentials
   - Verify network connectivity
   - Ensure database is running

2. **Redis connection failed:**
   - Check Redis address
   - Verify Redis is running
   - Check firewall rules

3. **Rate limiting not working:**
   - Verify Redis connection
   - Check rate limit configuration
   - Monitor Redis memory usage

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=debug
```

### Performance Tuning

1. **Database:**
   - Optimize queries
   - Add indexes
   - Use connection pooling

2. **Redis:**
   - Monitor memory usage
   - Configure eviction policies
   - Use Redis Cluster for large deployments

3. **Gateway:**
   - Tune Go runtime parameters
   - Monitor goroutine usage
   - Profile memory usage 