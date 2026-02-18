# Point of Sale System

## Overview
Admin panel with a Go backend API and Next.js frontend.

## Architecture
```
pointofsale/
├── backend/          # Go API server (see backend/CLAUDE.md)
├── frontend/         # Next.js admin panel (see frontend/CLAUDE.md)
├── e2e/              # End-to-end tests (Playwright + agent-browser)
│   ├── tests/        # Playwright test files by feature
│   │   ├── auth/     # Login, register, reset-password flows
│   │   ├── master/   # Category, product, supplier, rack CRUD
│   │   ├── transaction/ # Purchase order flows
│   │   └── settings/ # Users, roles, permissions
│   ├── helpers/      # Shared utilities (login, navigation, assertions)
│   └── playwright.config.ts
├── docker-compose.yml
└── prompts/          # Staged development prompt files
```

## Tech Stack
- **Frontend**: Next.js 16 (React 19), Tailwind CSS v4, Zustand v5, TypeScript
- **Backend**: Go 1.24, Chi v5, GORM, PostgreSQL 17, Redis 7
- **Auth**: Argon2id + JWT (access 15min, refresh 7d)
- **E2E Testing**: Playwright (automated) + agent-browser (AI-assisted QA)
- **Infra**: Docker Compose (backend, frontend, postgres, redis, mailpit)

## Running the Full Stack
```bash
docker compose up                     # all services
docker compose up backend             # backend only
cd frontend && npm run dev            # frontend dev server
cd backend && go test ./...           # backend tests (needs test DB)
```

## E2E Testing
**IMPORTANT**: When creating, editing, or fixing e2e tests, you MUST use the `/e2e-test` skill (`.claude/skills/e2e-test.md`) before writing any test code. This skill contains required patterns, debugging workflows, and conventions that ensure tests pass on the first few attempts.

## Running E2E Tests
```bash
# Via Docker (recommended — includes browser)
docker compose --profile test run --rm e2e

# Locally (needs Playwright browsers installed)
cd e2e && npx playwright install --with-deps chromium
cd e2e && npm test                    # headless
cd e2e && npm run test:headed         # with visible browser
cd e2e && npm run test:debug          # step-by-step debugger
cd e2e && npm run report              # view HTML report
```

## API Contract
- All routes under `/api/v1/` prefix
- Error format: `{"error": "message", "code": "CODE"}`
- Success format: `{"data": {...}, "message": "optional"}`
- Paginated lists: `{"data": [...], "meta": {"page", "pageSize", "totalItems", "totalPages"}}`

## Browser Automation

Use `agent-browser` for web automation. Run `agent-browser --help` for all commands.

Core workflow:
1. `agent-browser open <url>` - Navigate to page
2. `agent-browser snapshot -i` - Get interactive elements with refs (@e1, @e2)
3. `agent-browser click @e1` / `fill @e2 "text"` - Interact using refs
4. Re-snapshot after page changes

## Shared Conventions
- No shadcn-ui or external UI libraries — all components are custom-built
- Tailwind uses default color palette (custom palette planned for future)
- Super admin bypasses all permission checks
- Staged development via prompt files in `prompts/`
