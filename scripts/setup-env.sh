#!/bin/bash

# Tailscale MCP Server Environment Setup Script
# This script helps you set up your environment for local development

echo "🔧 Tailscale MCP Server Environment Setup"
echo "=========================================="
echo

# Check if .env already exists
if [ -f ".env" ]; then
  echo "⚠️  .env file already exists!"
  read -p "Do you want to overwrite it? (y/N): " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 0
  fi
fi

# Copy example file
echo "📋 Copying .env.example to .env..."
cp .env.example .env

# Create logs directory
echo "📁 Creating logs directory..."
mkdir -p logs

echo
echo "✅ Environment setup complete!"
echo
echo "📝 Next steps:"
echo "1. Edit .env file with your Tailscale credentials:"
echo "   - Get API key from: https://login.tailscale.com/admin/settings/keys"
echo "   - Set TAILSCALE_API_KEY and TAILSCALE_TAILNET"
echo
echo "2. Build the project:"
echo "   npm run build"
echo
echo "3. Test the server:"
echo "   node test-mcp-server.js"
echo
echo "4. Or run with MCP Inspector:"
echo "   npm run inspector"
echo
