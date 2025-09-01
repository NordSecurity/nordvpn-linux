# CI Workflows

## Overview

- **\core_ci.yml**: workflow for CLI/Tray/Daemon.
- **\gui_ci.yml**: workflow for GUI.

## Triggers

- **\core_ci.yml**: workflow for CLI/Tray/Daemon.
  - PRs, branch pushes, semver tags (`X.Y.Z`) - **for any file except `./gui/**`**
- **\gui_ci.yml**: workflow for GUI.
  - PRs, branch pushes, semver tags (`X.Y.Z`) - **for any file (including core files)**

## Rules

- Any non-`gui/**` change -> core + GUI.
- `gui/**`-only change -> GUI only.
