#!/bin/bash
set -euxo

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

mkdir -p $SCRIPT_DIR/configs/templates/ovpn/1.0
mkdir -p $SCRIPT_DIR/configs/templates/ovpn_xor/1.0
mkdir -p $SCRIPT_DIR/configs/dns

FILE1="configs/templates/ovpn/1.0/template.xslt"
FILE2="configs/templates/ovpn_xor/1.0/template.xslt"
FILE3="configs/dns/cybersec.json"

curl https://downloads.nordcdn.com/$FILE1 > $SCRIPT_DIR/$FILE1
curl https://downloads.nordcdn.com/$FILE2 > $SCRIPT_DIR/$FILE2
curl https://downloads.nordcdn.com/$FILE3 > $SCRIPT_DIR/$FILE3
