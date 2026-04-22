# /pipeline

Runs the full role-based development pipeline for: $ARGUMENTS

## Stages

Execute these stages **sequentially**. Each stage hands off via
files under `docs/pipeline/<current-branch>/`.

1. **architect** — explore the codebase, produce a plan, write it
   to `docs/pipeline/<branch>/plan.md`, then STOP and wait
2. **[HUMAN APPROVAL]** — present the plan and wait for the user
   to reply `approved` before continuing
3. **engineer** — read the approved plan, implement all components,
   write `docs/pipeline/<branch>/implementation.md`
4. **simplifier** — read implementation notes, simplify the code,
   write `docs/pipeline/<branch>/simplification.md`
5. **reviewer** — read all prior docs, review the code, write
   `docs/pipeline/<branch>/review.md`
6. **qa** — read all prior docs, write and run tests, write
   `docs/pipeline/<branch>/qa-report.md`

## Handoff Files (per branch)

All files live under `docs/pipeline/<branch>/` — auto-namespaced
to the current git branch so parallel worktrees never collide:

- `plan.md`             - written by architect, read by engineer
- `implementation.md`   - written by engineer, read by all after
- `simplification.md`   - written by simplifier, read by reviewer
- `review.md`           - written by reviewer, read by qa
- `qa-report.md`        - written by qa, final output

## Notes

- Do not proceed past stage 1 without explicit user approval
- If reviewer verdict is `NEEDS REVISION`, surface it to the user
  before running QA
- Each agent detects the branch via `git rev-parse --abbrev-ref HEAD`
- Works correctly in git worktrees — each worktree runs an
  independent pipeline
