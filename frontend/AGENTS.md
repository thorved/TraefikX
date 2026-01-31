# Frontend (Next.js)

Next.js 16 + React 19 + TypeScript + Tailwind CSS v4 + shadcn/ui

## Structure

```
src/
├── app/                    # Next.js App Router
│   ├── layout.tsx          # Root layout with providers
│   ├── page.tsx            # Home (redirects to dashboard)
│   ├── login/              # Login page
│   ├── dashboard/          # Dashboard
│   ├── users/              # User management
│   ├── profile/            # User profile
│   └── auth/oidc/callback/ # OIDC callback
├── components/
│   ├── ui/                 # shadcn/ui components
│   ├── layout/             # Layout components (sidebar, header)
│   └── providers/          # Context providers
├── contexts/
│   └── AuthContext.tsx     # Auth state
├── hooks/
│   ├── use-auth.ts         # Auth hook
│   └── use-users.ts        # Users hook
├── lib/
│   ├── api.ts              # Axios client + API functions
│   ├── utils.ts            # cn() and utilities
│   └── auth.ts             # Auth utilities
└── types/
    └── index.ts            # TypeScript types
```

## Patterns

### Components
- **Server Components by default** (no 'use client')
- Use `'use client'` for:
  - Event handlers (onClick, onSubmit)
  - React hooks (useState, useEffect, useContext)
  - Browser APIs (localStorage, window)
- shadcn/ui components in `components/ui/`

### Styling
- Tailwind CSS v4 with `@import "tailwindcss"` in globals.css
- Use `cn()` utility from `lib/utils.ts` for conditional classes
- Dark mode via `next-themes` ThemeProvider
- Icons from `lucide-react`

### State Management
- React Context API for global state (AuthContext)
- Custom hooks for data fetching (use-users.ts)
- Local state with useState for form inputs

### API Calls
- Use `api.ts` Axios instance (handles auth headers + token refresh)
- API functions exported as objects: `authApi.login()`, `usersApi.listUsers()`
- Types in `types/index.ts`

### shadcn/ui Components

Add new components:
```bash
npx shadcn@latest add button card input label dialog
```

Used components:
- **Layout**: sheet, separator, dropdown-menu
- **Forms**: input, label, button, select, switch
- **Data**: card, table, badge, avatar, skeleton, tabs
- **Feedback**: dialog, sonner

## Commands

```bash
npm run dev         # Dev server :3000
npm run build       # Production build
npm run lint        # Biome linter
npm run format      # Biome formatter
```

## Key Files

- `next.config.ts` - Static export + API proxy rewrites
- `lib/api.ts` - Axios instance with interceptors
- `contexts/AuthContext.tsx` - Auth state management
- `app/layout.tsx` - Root layout with ThemeProvider + AuthProvider

## TypeScript

- Strict mode enabled
- Path alias: `@/*` maps to `src/*`
- Types in `types/index.ts`: User, AuthResponse, ApiError
