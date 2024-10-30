#!/bin/bash

go mod edit -require=moose/events@v0.0.0
go mod edit -require=moose/worker@v0.0.0
go mod edit -replace=moose/events=./third-party/moose-events/nordvpnapp/moosenordvpnappgo
go mod edit -replace=moose/worker=./third-party/moose-worker/mooseworkergo
function revert_moose_patch {
    go mod edit -droprequire=moose/events
    go mod edit -droprequire=moose/worker
    go mod edit -dropreplace=moose/events
    go mod edit -dropreplace=moose/worker
}
trap revert_moose_patch EXIT