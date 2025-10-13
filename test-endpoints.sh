#!/bin/bash

echo "=== Testing mc-tool Web UI ==="
echo ""

# Start server in background
./mc-tool-portable web &
WEB_PID=$!
echo "Started web server (PID: $WEB_PID)"

# Wait for server to start
sleep 3

echo ""
echo "=== Testing /api/aliases endpoint ==="
curl -s http://localhost:8080/api/aliases
echo ""

echo ""
echo "=== Testing /api/aliases-stats endpoint ==="
curl -s http://localhost:8080/api/aliases-stats
echo ""

echo ""
echo "=== Stopping web server ==="
kill $WEB_PID 2>/dev/null
wait $WEB_PID 2>/dev/null

echo ""
echo "=== Test Complete ==="
