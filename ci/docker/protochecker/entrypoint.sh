#!/bin/sh
set -e

# take ownership
chown "${UID:-1000}:${GID:-1000}" /repo

# run the clone as the specified user.
su-exec "${UID:-1000}:${GID:-1000}" git clone /src /repo
