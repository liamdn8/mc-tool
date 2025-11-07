#!/bin/bash

echo "Starting mc-tool web server..."
./mc-tool-portable web &
WEB_PID=$!

sleep 3

echo ""
echo "Testing /api/aliases-stats endpoint..."
curl -s http://localhost:8080/api/aliases-stats

echo ""
echo ""
echo "Testing /api/aliases endpoint..."
curl -s http://localhost:8080/api/aliases

echo ""
echo ""
echo "Stopping web server (PID: $WEB_PID)..."
kill $WEB_PID 2>/dev/null
sleep 1

echo "Done!"
