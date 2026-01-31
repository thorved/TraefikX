# TraefikX

A modern Traefik GUI with user management, OIDC authentication, and proxy management capabilities.

## Features

- **User Management**: Admin interface to create, update, and manage users
- **Role-Based Access Control**: Admin and User roles with different permissions
- **Multiple Authentication Methods**:
  - Password-based authentication
  - OIDC (Pocket ID) integration
  - Account linking/unlinking capabilities
- **Secure by Default**:
  - JWT token authentication
  - Password complexity requirements
  - Session management
  - CSRF protection
- **Modern UI**: Built with Vue 3, shadcn/ui components, and Tailwind CSS

## Tech Stack

- **Backend**: Go 1.23, Gin, GORM, SQLite
- **Frontend**: Vue 3, TypeScript, Vite, Tailwind CSS, shadcn/ui
- **Authentication**: JWT, OAuth2/OIDC, bcrypt

## Quick Start

### Prerequisites

- Go 1.23+
- Node.js 20+
- SQLite

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd TraefikX
   ```

2. **Backend Setup**
   ```bash
   cd backend
   
   # Copy environment file
   cp .env.example .env
   
   # Install dependencies
   go mod tidy
   
   # Run the server
   go run cmd/api/main.go
   ```

3. **Frontend Setup** (in a new terminal)
   ```bash
   cd frontend
   
   # Install dependencies
   npm install
   
   # Run development server
   npm run dev
   ```

4. **Access the application**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - Default admin: `admin@traefikx.local` / `changeme`

### Production Deployment

Using Docker:

```bash
# Build and run
docker-compose up -d

# Access at http://localhost:8080
```

## Configuration

### Environment Variables

Create a `.env` file in the `backend` directory:

```env
# Server
PORT=8080
JWT_SECRET=your-super-secret-key-min-32-characters
ENV=development

# Database
DATABASE_PATH=./data/traefikx.db

# Security
BCRYPT_COST=12
ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=168h

# OIDC - Pocket ID Configuration
OIDC_ENABLED=true
OIDC_PROVIDER_NAME=Pocket ID
OIDC_ISSUER_URL=https://pocketid.example.com
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URL=http://localhost:8080/api/auth/oidc/callback
OIDC_SCOPES=openid,profile,email

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:8080

# Default Admin
DEFAULT_ADMIN_EMAIL=admin@traefikx.local
DEFAULT_ADMIN_PASSWORD=changeme
```

### OIDC Configuration (Pocket ID)

1. Create a new OIDC application in your Pocket ID instance
2. Set the redirect URL to: `http://localhost:8080/api/auth/oidc/callback`
3. Copy the Client ID and Client Secret to your `.env` file
4. Enable OIDC by setting `OIDC_ENABLED=true`

## API Endpoints

### Authentication
- `POST /api/auth/login` - Password login
- `POST /api/auth/logout` - Logout
- `POST /api/auth/refresh` - Refresh access token
- `GET /api/auth/oidc` - Initiate OIDC login
- `GET /api/auth/oidc/callback` - OIDC callback
- `GET /api/auth/oidc/status` - Get OIDC configuration status
- `GET /api/auth/me` - Get current user
- `PUT /api/auth/password` - Change password

### Users (Admin only)
- `GET /api/users` - List all users
- `POST /api/users` - Create user
- `GET /api/users/:id` - Get user
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user
- `POST /api/users/:id/reset-password` - Reset user password

## Project Structure

```
TraefikX/
├── backend/
│   ├── cmd/api/           # Main application entry
│   ├── internal/
│   │   ├── auth/          # JWT and OIDC logic
│   │   ├── config/        # Configuration management
│   │   ├── database/      # Database connection and migrations
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # Authentication middleware
│   │   └── models/        # Database models
│   ├── .env.example
│   ├── go.mod
│   └── Dockerfile.dev
├── frontend/
│   ├── src/
│   │   ├── api/           # API client
│   │   ├── components/    # UI components
│   │   ├── router/        # Vue Router config
│   │   ├── stores/        # Pinia stores
│   │   ├── views/         # Page components
│   │   └── types/         # TypeScript types
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
├── docker-compose.yml
└── Dockerfile
```

## Security Considerations

1. **Change default admin password immediately** after first login
2. Use strong JWT_SECRET (at least 32 characters)
3. Enable HTTPS in production
4. Use environment-specific CORS settings
5. Regularly rotate OIDC client secrets
6. Backup your SQLite database regularly

## Development

### Running Tests

Backend:
```bash
cd backend
go test ./...
```

Frontend:
```bash
cd frontend
npm run test
```

### Hot Reload

For development with hot reload:
```bash
cd backend
docker-compose --profile dev up traefikx-dev
```

## License

[Your License Here]