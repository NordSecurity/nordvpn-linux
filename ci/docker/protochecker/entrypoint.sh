#!/bin/sh
set -e

# take ownership
chown "$(id -u):$(id -g)" /repo

# run the clone as the specified user.
su-exec "$(id -u):$(id -g)" git clone /src /repo
