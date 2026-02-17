# Skill: Commit Changes

Commits staged and unstaged changes to the local git repository.

## When to Use
When the user asks to commit changes, save progress, or create a git commit.

## Steps

### 1. Check Status
- Run `git status` to see all modified, staged, and untracked files
- Run `git diff` to review unstaged changes
- Run `git diff --cached` to review already-staged changes
- Run `git log --oneline -5` to see recent commit message style

### 2. Stage Files
- Stage relevant files by name (e.g., `git add backend/handlers/foo.go`)
- Do NOT use `git add -A` or `git add .` — always add specific files
- Do NOT stage files that contain secrets (.env, credentials, etc.)
- Ask the user if unsure which files to include

### 3. Write Commit Message
- Analyze all staged changes to understand what was done
- Use the conventional commit format matching recent repo history
- Keep the subject line under 72 characters
- Focus on the "why" not the "what"
- End the message with: `Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>`

### 4. Commit
- Create the commit using a HEREDOC for the message:
  ```
  git commit -m "$(cat <<'EOF'
  feat(scope): short description

  Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
  EOF
  )"
  ```
- Run `git status` after to verify success

## Key Rules
- NEVER amend a previous commit unless explicitly asked
- NEVER use `--no-verify` to skip hooks
- If a pre-commit hook fails, fix the issue and create a NEW commit
- If there are no changes to commit, tell the user — do not create an empty commit
