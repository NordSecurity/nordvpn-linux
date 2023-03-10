#!/bin/bash
mkdir -p  /root/.pulp/ && chmod 700 /root/.pulp

cat <<EOF > /root/.pulp/admin.conf
[server]
host=$PULP_HOST
verify_ssl=$PULP_SSL
[auth]
username=$PULP_USER
password=$PULP_PASS
EOF

chmod 600 /root/.pulp/admin.conf
exec "$@";
