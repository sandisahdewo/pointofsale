# Skill: Add UI Component

Creates a new reusable UI component in the `components/ui/` directory.

## When to Use
When the user asks to add a new shared/reusable UI component.

## File Location
`frontend/src/components/ui/{ComponentName}.tsx`

## Conventions
- `'use client'` directive at top
- Default export with PascalCase name matching filename
- TypeScript interface for props: `{ComponentName}Props`
- Extend native HTML element props where applicable (e.g., `React.ButtonHTMLAttributes<HTMLButtonElement>`)
- Support `className` prop for overrides (append to default classes)
- Use Tailwind CSS for all styling — no CSS modules, no inline styles
- Color palette: blue-600 for primary, gray-300 borders, red-500/600 for errors/danger
- Focus states: `focus:outline-none focus:ring-2 focus:ring-blue-500`
- Error states: `border-red-500 focus:ring-red-500`
- Label pattern: optional `label` prop renders a `<label>` above the input
- Error pattern: optional `error` string renders `<p className="mt-1 text-sm text-red-600">`
- Size variants: sm/md/lg where applicable
- No external dependencies — keep components self-contained

## Examples of Existing Components
Reference these for consistent patterns:
- Simple input: `components/ui/Input.tsx`
- Button with variants: `components/ui/Button.tsx`
- Modal: `components/ui/Modal.tsx`
- Complex component: `components/ui/Table.tsx`

## After Creating
Update `CLAUDE.md` UI Components table if the component is significant.
