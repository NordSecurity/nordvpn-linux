#!/bin/bash

sudo umount -v /etc/resolv.conf
echo nameserver 1.1.1.1 | sudo tee /etc/resolv.conf
exec $@