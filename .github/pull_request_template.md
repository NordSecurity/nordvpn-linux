# Pull Request Template 🚀

Thanks for opening a PR - every small improvement helps. Use this guide to keep reviews smooth and the codebase healthy.

## 1. Status & Draft 📝
Keep the PR in **Draft** while you are still iterating (build, unit tests, lint, CI). Move to Ready for Review after a quick self‑scan for obvious issues.

## 2. Description 📣
Give reviewers the essentials:
* Current situation / problem.
* What changed (modules, files, behaviors).
* Why it changed (link issue: `LVPN-####` if available).
* Expected outcome / improvement.

## 3. Size & Scope 📦
Prefer focused PRs. If you must land something large, consider splitting it into logical parts or a temporary feature branch. If major follow‑up work appears mid‑review, open a new PR rather than stacking huge diffs. Also, when a substantial new change lands, **reset approvals** and ask for fresh reviews. 🔄

## 4. Tests 🧪
Add or update tests for new logic, edge cases, and error paths. Skipping tests should be rare and explicitly justified. Adapting existing tests is fine - explain briefly if coverage shifts.

## 5. Ownership & Hygiene 🔧
You own the PR until merge:
* Keep CI green (fix red builds early).
* Resolve or explicitly defer (with issue link) all comments.
* Rebase / merge main to clear conflicts early.
* Avoid stale PRs lingering without progress.
* Consider using labels and assign yourself to the PR.

### Quick Tips ✨
* Separate mechanical changes (format, rename) from logic changes when possible.
* Comment tricky decisions directly in code; reviewers appreciate context.
* Small, frequent PRs > giant multi‑week diffs.
* If the change affects any user‑visible interface, add a screenshot or short clip.

### Cleanup 🧹
Once you fill in your PR description above, feel free to remove unused guidance sections from this template in the PR body to keep things lean.

Thanks again for contributing! 🙌