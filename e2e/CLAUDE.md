# Point of Sale - E2E Tests

## Tech Stack
- **Test Framework**: Playwright (automated CI/CD tests)
- **Browser Automation**: agent-browser (AI-assisted QA during development)
- **Language**: TypeScript
- **Docker Image**: `mcr.microsoft.com/playwright:v1.58.2-noble`

## Project Structure

```
e2e/
├── tests/                  # Playwright test files by feature
│   ├── auth/               # Login, register, reset-password flows
│   ├── master/             # Category, product, supplier, rack CRUD
│   ├── transaction/        # Purchase order flows
│   └── settings/           # Users, roles, permissions
├── helpers/                # Shared utilities
│   └── auth.ts             # Login helper
├── playwright.config.ts    # Playwright configuration
├── tsconfig.json
├── Dockerfile
└── package.json
```

## Running Tests

```bash
# Via Docker (recommended — browser pre-installed)
docker compose --profile test run --rm e2e

# Locally (needs browser installed first)
npx playwright install --with-deps chromium
npm test                    # headless
npm run test:headed         # with visible browser
npm run test:debug          # step-by-step debugger
npm run report              # view HTML report
```

## Browser Automation (agent-browser)

Use `agent-browser` for interactive QA during development. Run `agent-browser --help` for all commands.

Core workflow:
1. `agent-browser open <url>` - Navigate to page
2. `agent-browser snapshot -i` - Get interactive elements with refs (@e1, @e2)
3. `agent-browser click @e1` / `fill @e2 "text"` - Interact using refs
4. Re-snapshot after page changes

## Conventions

### Test File Naming
```
tests/{feature}/{feature}.spec.ts
# Examples:
# tests/auth/login.spec.ts
# tests/master/category.spec.ts
# tests/transaction/purchase-order.spec.ts
```

### Test Structure
```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
  test('should do something expected', async ({ page }) => {
    await page.goto('/some-page');
    // arrange, act, assert
  });
});
```

### What to Test Per Page
1. **Navigation** — page loads, correct URL, title/heading visible
2. **CRUD operations** — create, read, update, delete via UI
3. **Form validation** — required fields, invalid input, error messages
4. **Auth flows** — login required, redirect to login, permission denied
5. **User feedback** — toast notifications, success/error messages
6. **Edge cases** — empty states, pagination, search/filter

### Helpers
- Use `helpers/auth.ts` for login — avoid repeating login steps in every test
- Add shared utilities to `helpers/` (e.g., navigation, data seeding)

### Environment
- `BASE_URL` defaults to `http://localhost:3000` (overridden to `http://frontend:3000` in Docker)
- `CI=true` in Docker — enables stricter settings (no `test.only`, retries, HTML reporter)
