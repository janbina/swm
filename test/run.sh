#!/bin/bash

DISPLAY=:111

echo "Building swm, swmctl and test"
go build github.com/janbina/swm/cmd/swm
go build github.com/janbina/swm/cmd/swmctl
go build github.com/janbina/swm/test

echo "Starting xvfb and swm"

Xvfb $DISPLAY -screen 0 1024x768x16 &
XVFB_PID=$!
sleep 1
./swm  > /dev/null 2>&1 &
sleep 1

echo "=================================================="
./test
EXIT_CODE=$?

kill -15 $XVFB_PID

rm swm swmctl test

exit $EXIT_CODE
