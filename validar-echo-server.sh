#!/bin/bash

TEST_MESSAGE="hello echoer"
PORT=12345

RESPONSE=$(docker run --rm --network tp0_testing_net --name netcat ubuntu:18.04 bash -c "
    apt-get update > /dev/null 2>&1 && \
    apt-get install -y netcat-traditional > /dev/null 2>&1 && \
    echo \"$TEST_MESSAGE\" | nc -w 1 "server" $PORT
")

if [ "$RESPONSE" = "$TEST_MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
