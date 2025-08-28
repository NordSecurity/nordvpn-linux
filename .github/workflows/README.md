# CI Workflows

## Overview

- **ci-orchestrator.yml**: decides whether to run `_core_ci.yml`, `_gui_ci.yml`,
  or both.
- **\_core_ci.yml**: reusable workflow for CLI/Tray/Daemon.
- **\_gui_ci.yml**: reusable workflow for GUI.

## Triggers

- PRs, branch pushes, semver tags (`X.Y.Z`)

## Rules

- Any non-`gui/**` change -> core + GUI.
- `gui/**`-only change -> GUI only.
