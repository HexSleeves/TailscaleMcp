#!/bin/bash

echo "Testing signal handling with stdin kept open..."

# Create a named pipe to keep stdin open
mkfifo test_pipe

# Start the server with the pipe as stdin
./tailscale-mcp-server/tailscale-mcp-server serve --mode=stdio --verbose < test_pipe &
SERVER_PID=$!

echo "Server started with PID: $SERVER_PID"
sleep 2

echo "Sending SIGINT to server..."
kill -INT $SERVER_PID

# Wait for the server to terminate
wait $SERVER_PID
EXIT_CODE=$?

echo "Server terminated with exit code: $EXIT_CODE"

# Clean up
rm -f test_pipe

if [ $EXIT_CODE -eq 0 ] || [ $EXIT_CODE -eq 1 ]; then
    echo "✅ SUCCESS: Server responded to SIGINT and terminated gracefully"
else
    echo "❌ FAILURE: Server did not terminate properly (exit code: $EXIT_CODE)"
fi
