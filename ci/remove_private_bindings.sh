#!/bin/bash

bindings_name=$1

go mod edit -droprequire="$bindings_name"
go mod edit -dropreplace="$bindings_name"