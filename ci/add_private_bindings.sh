#!/bin/bash

bindings_name=$1
bindings_path=$2

go mod edit -require="$bindings_name"@v0.0.0
go mod edit -replace="$bindings_name"="$bindings_path"
