#!/usr/bin/env bash
set -euxo pipefail

# Usage: slackpost "<webhook_url>" "<channel>" "<username>" "<message>"
# also (echo $RANDOM; echo $RANDOM) |slackpost "<channel>" "<username>"

# ------------
webhook_url=$1
if [[ $webhook_url == "" ]]
then
        echo "No webhook_url specified"
        exit 1
fi

# ------------
shift
channel=$1
if [[ $channel == "" ]]
then
        echo "No channel specified"
        exit 1
fi

# ------------
shift
username=$1
if [[ $username == "" ]]
then
        echo "No username specified"
        exit 1
fi

# ------------
shift

text=$*

if [[ $text == "" ]]
then
while IFS= read -r line; do
  #printf '%s\n' "$line"
  text="$text$line\n"
done
fi


if [[ $text == "" ]]
then
        echo "No text specified"
        exit 1
fi

escapedText=$(echo "${text}" | sed 's/"/\"/g' | sed "s/'/\'/g" )

json="{\"channel\": \"$channel\", \"username\":\"$username\", \"icon_emoji\":\":penguin:\", \"attachments\":[{\"color\":\"good\",\"text\": \"$escapedText\",\"mrkdwn_in\": [\"fields\"]}]}"

curl -s -d "payload=$json" "$webhook_url"
