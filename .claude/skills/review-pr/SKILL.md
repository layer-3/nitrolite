---
name: review-pr
description: Thorough code review of current changes or a specific PR
allowed-tools: Bash, Read, Grep, Glob
---

Review changes for: $ARGUMENTS

1. **Get the diff:**
   - If $ARGUMENTS is a PR number: `gh pr diff $ARGUMENTS`
   - If no arguments: `git diff` and `git diff --staged` for uncommitted changes

2. **For each changed file, check:**
   - **Correctness:** Does the logic do what it claims? Any off-by-one errors, race conditions, or wrong assumptions?
   - **Types:** Are TypeScript types accurate and strict? Are Go errors properly handled?
   - **Security:** No exposed secrets, no injection vectors, no reentrancy (contracts), no XSS
   - **Tests:** Are new functions tested? Are edge cases and error paths covered?
   - **Exports:** If public API changed, is the barrel `index.ts` updated? (sdk-compat: no SSR-unsafe re-exports)
   - **Docs:** If behavior changed, are README/CLAUDE.md/migration docs updated?
   - **Style:** Consistent with existing patterns in the file and project conventions

3. **Categorize findings:**
   - **CRITICAL:** Must fix before merge (bugs, security issues, data loss risk)
   - **WARNING:** Should fix (missing tests, poor error handling, incomplete docs)
   - **SUGGESTION:** Nice to have (naming improvements, minor refactors, style tweaks)

4. **Give an overall verdict:** APPROVE / REQUEST CHANGES / COMMENT
