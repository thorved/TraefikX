# Build stage for frontend
FROM node:alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ .
RUN npm run build

# Build stage for backend
FROM golang:alpine AS backend-builder


WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN GOOS=linux go build -o api cmd/api/main.go

# Final stage
FROM alpine:latest
WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/backend/api .

# Copy frontend build
COPY --from=frontend-builder /app/frontend/out ./frontend/out

# Create data directory
RUN mkdir -p /app/data

# Environment variables
ENV PORT=8080
ENV ENV=production
ENV DATABASE_PATH=/app/data/traefikx.db

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api"]