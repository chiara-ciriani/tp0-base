#!/bin/bash

PORT=12345
MESSAGE="Hello world!"
TIMEOUT=5
SERVER_CONTAINER="server"

RESPONSE=$(docker run --rm --network tp0_testing_net alpine:latest sh -c "echo $MESSAGE | nc -w $TIMEOUT $SERVER_CONTAINER $PORT")

if [ "$RESPONSE" == "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi
