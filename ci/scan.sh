#!/bin/bash
set -euox

gosec -quiet -exclude-dir=third-party ./...
