#!/bin/bash

echo "Testing signal handling in HTTP mode..."

# Start the server in HTTP mode
./tailscale-mcp-server/tailscale-mcp-server serve --mode=http --port=8080 --verbose &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
sleep 3

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
