#!/bin/bash

set -euxo

/usr/sbin/sshd
cd /
python3 /server.py

sleep infinity
