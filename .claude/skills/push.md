# Skill: Push to Remote

Pushes local commits to the remote git repository.

## When to Use
When the user asks to push changes, sync with remote, or upload commits.

## Steps

### 1. Pre-Push Checks
- Run `git status` to confirm there are no uncommitted changes the user may want to include
- Run `git log --oneline origin/HEAD..HEAD` to show which commits will be pushed (if tracking branch exists)
- Check the current branch name with `git branch --show-current`

### 2. Confirm with User
- Show the user which branch and how many commits will be pushed
- If pushing to `main` or `master`, warn the user and ask for confirmation
- If there are uncommitted changes, ask if they want to commit first

### 3. Push
- Push with upstream tracking: `git push -u origin <branch>`
- NEVER use `--force` or `--force-with-lease` unless explicitly asked
- If push is rejected (non-fast-forward), inform the user and suggest pulling first — do NOT force push

## Key Rules
- NEVER force push to main/master
- NEVER push without showing the user what will be pushed first
- If the remote rejects the push, explain why and let the user decide next steps
