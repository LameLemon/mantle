#!/usr/bin/env bash

set -e
set -x

./node_modules/.bin/jshint ./www/chat/js/*.js

go test

go build
./mantle \
