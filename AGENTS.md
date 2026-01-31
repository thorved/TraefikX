# TraefikX - Opencode Agent Configuration

## Project Overview
TraefikX is a modern Traefik GUI with user management, OIDC authentication, and proxy management capabilities.

## Tech Stack
- **Backend**: Go 1.23 (Gin, GORM, SQLite)
- **Frontend**: Next.js 16 + React 19 + TypeScript + Tailwind CSS v4 + shadcn/ui
- **Authentication**: JWT, OAuth2/OIDC (Pocket ID)
- **State Management**: React Context API
- **HTTP Client**: Axios with interceptors
- **Icons**: Lucide React

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
  ├─ app/                       # Next.js App Router
  │   ├─ login/
  │   │   └─ page.tsx           # Login page
  │   ├─ dashboard/
  │   │   └─ page.tsx           # Dashboard
  │   ├─ users/
  │   │   └─ page.tsx           # User management
  │   ├─ profile/
  │   │   └─ page.tsx           # User profile
  │   ├─ auth/
  │   │   └─ oidc/
  │   │       └─ callback/
  │   │           └─ page.tsx   # OIDC callback
  │   ├─ layout.tsx             # Root layout with providers
  │   ├─ page.tsx               # Home (redirects to dashboard)
  │   └─ globals.css            # Global styles
  ├─ components/
  │   ├─ ui/                    # shadcn/ui components
  │   │   ├─ button.tsx
  │   │   ├─ card.tsx
  │   │   ├─ input.tsx
  │   │   ├─ label.tsx
  │   │   ├─ dialog.tsx
  │   │   ├─ dropdown-menu.tsx
  │   │   ├─ table.tsx
  │   │   ├─ badge.tsx
  │   │   ├─ avatar.tsx
  │   │   ├─ toast.tsx
  │   │   ├─ sonner.tsx
  │   │   ├─ select.tsx
  │   │   ├─ switch.tsx
  │   │   ├─ tabs.tsx
  │   │   ├─ separator.tsx
  │   │   ├─ skeleton.tsx
  │   │   └─ sheet.tsx
  │   ├─ layout/
  │   │   ├─ sidebar.tsx        # Navigation sidebar
  │   │   ├─ header.tsx         # Top header with user menu
  │   │   └─ protected-layout.tsx # Protected route wrapper
  │   └─ providers/
  │       └─ auth-provider.tsx  # Authentication context provider
  ├─ contexts/
  │   └─ auth-context.tsx       # React Context for auth state
  ├─ lib/
  │   ├─ api.ts                 # Axios API client
  │   ├─ utils.ts               # Utility functions (cn, etc.)
  │   └─ auth.ts                # Auth utilities
  ├─ types/
  │   └─ index.ts               # TypeScript type definitions
  ├─ hooks/
  │   ├─ use-auth.ts            # Auth hook
  │   └─ use-users.ts           # Users data hook
  └─ lib/
      ├─ api.ts                 # Axios API client
      └─ utils.ts               # Utility functions
```

## Key Files to Know

### Configuration
- `backend/.env` - Environment variables (copy from .env.example)
- `docker-compose.yml` - Docker deployment config
- `Dockerfile` - Multi-stage production build
- `frontend/next.config.ts` - Next.js config with static export & API proxy rewrites

### Database Models
- **User**: email, password, role (admin/user), OIDC fields, timestamps
- **Session**: user_id, token, expires_at

### API Endpoints
- Auth: `/api/auth/*` - login, logout, refresh, OIDC, me
- Users: `/api/users/*` - CRUD operations (admin only)

### Frontend Routes
- `/login` - Login page
- `/dashboard` - Dashboard
- `/users` - User management (admin only)
- `/profile` - User profile
- `/auth/oidc/callback` - OIDC callback

### Proxy Configuration (Next.js)
The frontend uses Next.js rewrites to proxy API requests to the backend:
```typescript
// next.config.ts
async rewrites() {
  return [
    {
      source: '/api/:path*',
      destination: 'http://localhost:8080/api/:path*',
    },
  ]
}
```

## Development Workflow

### Running Locally
1. Backend: `cd backend && go run cmd/api/main.go` (runs on :8080)
2. Frontend: `cd frontend && npm run dev` (runs on :3000)
3. Access: http://localhost:3000

### Default Credentials
- Email: admin@traefikx.local
- Password: changeme

### Building for Production
```bash
# Frontend

# Install dependencies and shadcn components
npm install
npx shadcn@latest add button card input label dialog dropdown-menu table badge avatar sonner select switch tabs separator skeleton sheet

# Build static export
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
- **Backend**: REST API with Gin, GORM for database
- **Frontend**: React Server Components + Client Components, Context for state
- **Types**: Always use TypeScript interfaces
- **Error handling**: Return structured JSON errors
- **shadcn/ui**: Use `npx shadcn@latest add <component>` to add components

### Testing
- Backend: `go test ./...`
- Frontend: `npm run lint` (Biome linter)

### Common Tasks
1. Add new API endpoint → Create handler → Add route in main.go
2. Add new page → Create `page.tsx` in app directory
3. Add new component → Use `npx shadcn@latest add <component>`
4. Database changes → Update models → Auto-migrate on start

## Dependencies to Know

### Backend
- `gin-gonic/gin` - HTTP framework
- `golang-jwt/jwt/v5` - JWT tokens
- `gorm.io/gorm` - ORM
- `golang.org/x/crypto/bcrypt` - Password hashing
- `golang.org/x/oauth2` - OIDC support

### Frontend
- `next` (v16) + `react` (v19) + `react-dom` (v19) - Framework
- `tailwindcss` (v4) - Styling
- `@radix-ui/*` - Component primitives (via shadcn)
- `class-variance-authority` + `clsx` + `tailwind-merge` - Styling utilities
- `lucide-react` - Icons
- `axios` - HTTP client
- `zod` - Schema validation
- `sonner` - Toast notifications

## Next Steps for Proxy Management

When implementing Traefik proxy management:
1. Create models: Proxy, Router, Service, Middleware
2. Add handlers: CRUD for proxies
3. Create views: Proxy list, Proxy detail, Create proxy
4. Integrate with Traefik API (if using Traefik API) or file provider
5. Add validation for proxy configurations

## shadcn/ui Components Used

### Navigation
- `sheet` - Mobile sidebar
- `separator` - Dividers
- `dropdown-menu` - User menu

### Forms
- `input` - Text inputs
- `label` - Form labels
- `button` - Actions
- `select` - Dropdown selects
- `switch` - Toggles

### Data Display
- `card` - Content containers
- `table` - Data tables
- `badge` - Status indicators
- `avatar` - User avatars
- `skeleton` - Loading states
- `tabs` - Tabbed content

### Feedback
- `dialog` - Modals
- `sonner` - Toast notifications

### Layout
- All components support dark mode via Tailwind CSS
