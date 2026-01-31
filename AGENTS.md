# TraefikX Project

Modern Traefik GUI with user management, OIDC authentication, and proxy management.

## Architecture

- **Backend**: Go 1.25 + Gin + GORM + SQLite
- **Frontend**: Next.js 16 + React 19 + TypeScript + Tailwind CSS v4 + shadcn/ui
- **Auth**: JWT + OIDC (Pocket ID)

## Quick Start

```bash
# Backend
cd backend && go run cmd/api/main.go  # :8080

# Frontend  
cd frontend && npm run dev            # :3000
```

## Default Credentials
- Email: `admin@traefikx.local`
- Password: `changeme`

## Project Structure

```
backend/           # Go API
frontend/          # Next.js app
├── src/
│   ├── app/       # Next.js App Router
│   ├── components/# React components
│   ├── contexts/  # React Context
│   ├── hooks/     # Custom hooks
│   ├── lib/       # Utilities
│   └── types/     # TypeScript types
```

## Development Commands

- Backend: `go run cmd/api/main.go`, `go test ./...`
- Frontend: `npm run dev`, `npm run lint`, `npm run format`
- Docker: `docker-compose up -d`

## Key Patterns

- REST API with JSON responses
- JWT Bearer token auth (Authorization header)
- React Context for global state
- shadcn/ui components via Radix UI
- Server Components by default, 'use client' when needed
