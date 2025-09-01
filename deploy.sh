#!/bin/bash

# API Gateway Deployment Script
# This script helps deploy your API Gateway to various hosting platforms

set -e

echo "🚀 API Gateway Deployment Script"
echo "================================"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Set environment variables
export DB_PASSWORD=${DB_PASSWORD:-"your_secure_password_here"}
export JWT_SECRET=${JWT_SECRET:-"your_super_secret_jwt_key_here"}

echo "📋 Deployment Options:"
echo "1. Local Docker deployment"
echo "2. DigitalOcean Droplet"
echo "3. AWS EC2"
echo "4. Google Cloud Run"
echo "5. Heroku"
echo "6. Railway"

read -p "Choose deployment option (1-6): " choice

case $choice in
    1)
        echo "🐳 Deploying locally with Docker..."
        cd docker
        docker-compose -f docker-compose.prod.yml up -d
        echo "✅ Local deployment complete!"
        echo "🌐 Frontend: http://localhost"
        echo "🔧 Backend: http://localhost:8080"
        ;;
    2)
        echo "🌊 Deploying to DigitalOcean..."
        echo "Please follow these steps:"
        echo "1. Create a DigitalOcean account"
        echo "2. Create a new Droplet (Ubuntu 22.04)"
        echo "3. SSH into your Droplet"
        echo "4. Run: curl -fsSL https://get.docker.com -o get-docker.sh && sh get-docker.sh"
        echo "5. Install Docker Compose: sudo apt-get install docker-compose"
        echo "6. Clone your repository"
        echo "7. Run: cd docker && docker-compose -f docker-compose.prod.yml up -d"
        ;;
    3)
        echo "☁️ Deploying to AWS EC2..."
        echo "Please follow these steps:"
        echo "1. Create an AWS account"
        echo "2. Launch an EC2 instance (Amazon Linux 2)"
        echo "3. Configure security groups (ports 80, 443, 22)"
        echo "4. SSH into your instance"
        echo "5. Install Docker: sudo yum update -y && sudo yum install -y docker"
        echo "6. Start Docker: sudo systemctl start docker && sudo systemctl enable docker"
        echo "7. Install Docker Compose"
        echo "8. Clone your repository and deploy"
        ;;
    4)
        echo "🔧 Deploying to Google Cloud Run..."
        echo "Please follow these steps:"
        echo "1. Create a Google Cloud account"
        echo "2. Enable Cloud Run API"
        echo "3. Build and push your Docker images to Google Container Registry"
        echo "4. Deploy using: gcloud run deploy"
        ;;
    5)
        echo "🦄 Deploying to Heroku..."
        echo "Please follow these steps:"
        echo "1. Create a Heroku account"
        echo "2. Install Heroku CLI"
        echo "3. Create a new Heroku app"
        echo "4. Add PostgreSQL addon: heroku addons:create heroku-postgresql"
        echo "5. Add Redis addon: heroku addons:create heroku-redis"
        echo "6. Deploy: git push heroku main"
        ;;
    6)
        echo "🚂 Deploying to Railway..."
        echo "Please follow these steps:"
        echo "1. Create a Railway account"
        echo "2. Connect your GitHub repository"
        echo "3. Add environment variables in Railway dashboard"
        echo "4. Deploy automatically from Git"
        ;;
    *)
        echo "❌ Invalid option. Please choose 1-6."
        exit 1
        ;;
esac

echo ""
echo "🔐 Security Checklist:"
echo "✅ Change default passwords"
echo "✅ Set up SSL/TLS certificates"
echo "✅ Configure firewall rules"
echo "✅ Set up monitoring and logging"
echo "✅ Regular backups"
echo ""
echo "📚 Documentation:"
echo "Frontend: http://localhost"
echo "API Docs: http://localhost/docs"
echo "Health Check: http://localhost/health"
