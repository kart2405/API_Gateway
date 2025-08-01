#!/bin/bash

# Database initialization script for API Gateway

set -e

echo "Initializing database..."

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h localhost -p 5432 -U postgres; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is ready!"

# Create database if it doesn't exist
echo "Creating database if it doesn't exist..."
psql -h localhost -p 5432 -U postgres -d postgres -c "CREATE DATABASE apigateway;" || echo "Database already exists"

# Run migrations and create default user
echo "Running database migrations and creating default user..."
go run cmd/gateway/main.go --init-db

echo "Database initialization complete!" 