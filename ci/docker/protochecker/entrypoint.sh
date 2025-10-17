#!/bin/sh
set -eux

# take ownership
chown "${USERID:-1000}:${GID:-1000}" /repo

# run the clone as the specified user.
su-exec "${USERID:-1000}:${GID:-1000}" git -c safe.directory=/src clone /src /repo