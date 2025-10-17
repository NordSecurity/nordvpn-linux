#!/bin/bash

args="$*"
if [[ $args == "--no-headers -o comm 1" ]]; then
    echo init
else
    pso "$args"
fi