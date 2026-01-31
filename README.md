# Brunch Card Digital

## Overview
Brunch Card Digital is a digital loyalty card application designed to enhance customer engagement through a rewards system. Customers can earn stamps for their purchases and redeem rewards once they reach a certain threshold.

## Features
- **Digital Loyalty Cards**: Customers can create and manage their digital loyalty cards.
- **QR Code Generation**: Each card has a unique QR code for easy access and verification.
- **PostgreSQL Database**: The application uses PostgreSQL for data storage, ensuring reliability and performance.
- **Kubernetes Deployment**: The application is containerized and can be deployed on Kubernetes, making it scalable and easy to manage.

## Technologies Used
- **Go**: The backend is built using Go, providing a fast and efficient server.
- **Docker**: The application is containerized using Docker, allowing for easy deployment and management.
- **Kubernetes**: The application is deployed on a Kubernetes cluster, ensuring scalability and reliability.

## Getting Started
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/brunch-card-digital.git
   cd brunch-card-digital
   ```
2. Build the Docker image:
   ```bash
   make build
   ```
3. Deploy to Kubernetes:
   ```bash
   make deploy
   ```

## API Endpoints
- **Create Card**: `POST /api/v1/cards`
- **Get QR Code**: `GET /api/v1/qrcode?id={cardID}`

## License
This project is licensed under the MIT License.
