#!/usr/bin/env sh

if [ -z "$(git status -s uds.schema.json)" ]; then
    echo "Success!"
    exit 0
else
    git status uds.schema.json
    exit 1
fi
