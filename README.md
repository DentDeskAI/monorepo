# DentDesk 🦷

> Multi-tenant SaaS platform for dental clinics — Go + React MVP

## Stack

| Layer     | Technology                                          |
|-----------|-----------------------------------------------------|
| Backend   | Go 1.23, Gin, GORM, PostgreSQL                     |
| Frontend  | React 19, Vite, TypeScript, Tailwind, shadcn/ui    |
| Messaging | WhatsApp Cloud API                                  |
| AI        | Groq / OpenAI / Anthropic (OpenAI-compatible API)  |
| Auth      | JWT (HS256) with clinic_id embedded for multi-tenancy |
| Infra     | Docker Compose (Postgres + Redis)                  |

## Quick Start

### 1. Clone & configure

```bash
git clone https://github.com/your-org/dentdesk.git
cd dentdesk
cp .env.example .env
# Edit .env with your credentials
```

### 2. Start infrastructure

```bash
cd backend
docker-compose up -d postgres redis
```

### 3. Run backend

```bash
cd backend
go mod download
go run ./cmd/dentdesk
# → API on http://localhost:8080
```

### 4. Run frontend

```bash
cd frontend
npm install
npm run dev
# → UI on http://localhost:5173
```

## Project Structure

```
dentdesk/
├── backend/
│   ├── cmd/dentdesk/main.go       # Entry point & wiring
│   ├── internal/
│   │   ├── config/                # Viper config
│   │   ├── domain/                # GORM models (Clinic, Patient, Appointment, ...)
│   │   ├── handler/               # Gin HTTP handlers
│   │   ├── service/               # Business logic
│   │   ├── repository/            # GORM data access
│   │   ├── whatsapp/              # Cloud API client + types
│   │   ├── llm/                   # OpenAI-compatible LLM client
│   │   ├── middleware/            # Tenant + RBAC middleware
│   │   └── routes/                # Route registration
│   └── migrations/                # GORM AutoMigrate
│
└── frontend/
    └── src/
        ├── app/                   # Router + App root
        ├── components/layout/     # Sidebar, TopBar, AppShell
        ├── features/              # dashboard, dialogs, calendar, patients
        ├── hooks/                 # useAuth, useTenant
        ├── lib/                   # api.ts, auth.ts, utils.ts
        └── types/                 # domain.ts, api.ts
```

## API Endpoints

| Method | Path                       | Auth | Description               |
|--------|----------------------------|------|---------------------------|
| POST   | /api/v1/auth/register      | ❌   | Register clinic + admin   |
| POST   | /api/v1/auth/login         | ❌   | Login, get JWT            |
| GET    | /api/v1/auth/me            | ✅   | Current user              |
| GET    | /api/v1/patients           | ✅   | List patients (paginated) |
| POST   | /api/v1/patients           | ✅   | Create patient            |
| GET    | /api/v1/appointments       | ✅   | List appointments         |
| POST   | /api/v1/appointments       | ✅   | Create appointment        |
| GET    | /webhook/whatsapp          | ❌   | Meta webhook verification |
| POST   | /webhook/whatsapp          | ❌   | Inbound WhatsApp events   |

## Multi-tenancy

Every protected API request requires a JWT. The `TenantMiddleware` extracts
`clinic_id` from the token and injects it into the Gin context. All repository
queries include `WHERE clinic_id = ?` — data is physically shared in one
database but logically isolated per tenant.

## Adding shadcn/ui components

```bash
cd frontend
npx shadcn@latest add button dialog select
```

## Environment Variables

See `.env.example` for the full list. Key variables:

- `JWT_SECRET` — at least 32 random characters in production
- `WHATSAPP_*` — from Meta Developer Console
- `LLM_*` — Groq, OpenAI, or Anthropic credentials
