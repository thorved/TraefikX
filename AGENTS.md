# TraefikX - Opencode Agent Configuration

## Project Overview
TraefikX is a modern Traefik GUI with user management, OIDC authentication, and proxy management capabilities.

## Tech Stack
- **Backend**: Go 1.23 (Gin, GORM, SQLite)
- **Frontend**: Vue 3 + TypeScript + Vite + Tailwind CSS + shadcn/ui
- **Authentication**: JWT, OAuth2/OIDC (Pocket ID)

## Directory Structure

### Backend (/backend)
```
cmd/api/
  └─ main.go                    # Application entry point
internal/
  ├─ auth/
  │   ├─ jwt.go                 # JWT token generation/validation
  │   └─ oidc.go                # OIDC configuration and flow
  ├─ config/
  │   └─ config.go              # Environment configuration
  ├─ database/
  │   ├─ database.go            # DB connection & migrations
  │   └─ password.go            # Password hashing utilities
  ├─ handlers/
  │   ├─ auth.go                # Authentication handlers
  │   ├─ oidc.go                # OIDC callback handlers
  │   └─ user.go                # User CRUD handlers
  ├─ middleware/
  │   └─ auth.go                # JWT auth & role middleware
  └─ models/
      └─ user.go                # User & Session models
```

### Frontend (/frontend)
```
src/
  ├─ api/
  │   └─ client.ts              # Axios API client with interceptors
  ├─ components/
  │   └─ ui/                    # shadcn/ui components
  │       ├─ button/
  │       ├─ card/
  │       ├─ input/
  │       └─ label/
  ├─ router/
  │   └─ index.ts               # Vue Router with auth guards
  ├─ stores/
  │   └─ auth.ts                # Pinia auth store
  ├─ views/
  │   ├─ LoginView.vue          # Login with password + OIDC
  │   ├─ DashboardView.vue      # Main dashboard
  │   ├─ UsersView.vue          # User management (admin)
  │   ├─ ProfileView.vue        # User profile & account linking
  │   └─ OIDCCallbackView.vue   # OIDC callback handler
  ├─ types/
  │   └─ index.ts               # TypeScript type definitions
  └─ lib/
      └─ utils.ts               # Utility functions (cn, etc.)
```

## Key Files to Know

### Configuration
- `backend/.env` - Environment variables (copy from .env.example)
- `docker-compose.yml` - Docker deployment config
- `Dockerfile` - Multi-stage production build

### Database Models
- **User**: email, password, role (admin/user), OIDC fields, timestamps
- **Session**: user_id, token, expires_at

### API Endpoints
- Auth: `/api/auth/*` - login, logout, refresh, OIDC, me
- Users: `/api/users/*` - CRUD operations (admin only)

### Frontend Routes
- `/login` - Login page
- `/` - Dashboard
- `/users` - User management (admin only)
- `/profile` - User profile
- `/auth/oidc/callback` - OIDC callback

## Development Workflow

### Running Locally
1. Backend: `cd backend && go run cmd/api/main.go`
2. Frontend: `cd frontend && npm run dev`
3. Access: http://localhost:5173

### Default Credentials
- Email: admin@traefikx.local
- Password: changeme

### Building for Production
```bash
# Frontend
cd frontend && npm run build

# Backend
cd backend && go build -o api cmd/api/main.go

# Docker
docker-compose up -d
```

## Important Notes

### Security
- JWT secret must be 32+ characters
- Passwords require: 12+ chars, upper, lower, number, special
- Change default admin password immediately

### OIDC Setup (Pocket ID)
1. Set redirect URL: `http://localhost:8080/api/auth/oidc/callback`
2. Configure in `.env`
3. Enable with `OIDC_ENABLED=true`

### Code Patterns
- Backend: REST API with Gin, GORM for database
- Frontend: Composition API, Pinia for state, shadcn/ui components
- Types: Always use TypeScript interfaces in frontend
- Error handling: Return structured JSON errors

### Testing
- Backend: `go test ./...`
- Frontend: Manual testing via UI

### Common Tasks
1. Add new API endpoint → Create handler → Add route in main.go
2. Add new page → Create view → Add route in router
3. Database changes → Update models → Auto-migrate on start
4. New component → Add to components/ui/ → Use shadcn patterns

## Dependencies to Know

### Backend
- `gin-gonic/gin` - HTTP framework
- `golang-jwt/jwt/v5` - JWT tokens
- `gorm.io/gorm` - ORM
- `golang.org/x/crypto/bcrypt` - Password hashing
- `golang.org/x/oauth2` - OIDC support

### Frontend
- `vue@3` + `vue-router@4` - Framework & routing
- `pinia` - State management
- `axios` - HTTP client
- `tailwindcss` + `class-variance-authority` + `clsx` + `tailwind-merge` - Styling
- `lucide-vue-next` - Icons
- `radix-vue` - Component primitives

## Next Steps for Proxy Management

When implementing Traefik proxy management:
1. Create models: Proxy, Router, Service, Middleware
2. Add handlers: CRUD for proxies
3. Create views: Proxy list, Proxy detail, Create proxy
4. Integrate with Traefik API (if using Traefik API) or file provider
5. Add validation for proxy configurations
