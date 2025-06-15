#!/bin/bash

echo "Testing signal handling for tailscale-mcp-server..."

# Start the server in the background
./tailscale-mcp-server/tailscale-mcp-server serve --mode=stdio --verbose &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
sleep 2

echo "Sending SIGINT to server..."
kill -INT $SERVER_PID

# Wait for the server to terminate
wait $SERVER_PID
EXIT_CODE=$?

echo "Server terminated with exit code: $EXIT_CODE"

if [ $EXIT_CODE -eq 0 ] || [ $EXIT_CODE -eq 1 ]; then
    echo "✅ SUCCESS: Server responded to SIGINT and terminated gracefully"
else
    echo "❌ FAILURE: Server did not terminate properly (exit code: $EXIT_CODE)"
fi
