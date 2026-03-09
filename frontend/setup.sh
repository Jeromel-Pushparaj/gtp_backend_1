#!/bin/bash

echo "🚀 Setting up Slack Approval Workflow Frontend..."
echo ""

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Node.js version 18 or higher is required. Current version: $(node -v)"
    exit 1
fi

echo "✅ Node.js $(node -v) detected"
echo ""

# Check if package manager is available
if command -v pnpm &> /dev/null; then
    PKG_MANAGER="pnpm"
elif command -v yarn &> /dev/null; then
    PKG_MANAGER="yarn"
else
    PKG_MANAGER="npm"
fi

echo "📦 Using package manager: $PKG_MANAGER"
echo ""

# Install dependencies
echo "📥 Installing dependencies..."
$PKG_MANAGER install

if [ $? -ne 0 ]; then
    echo "❌ Failed to install dependencies"
    exit 1
fi

echo ""
echo "✅ Dependencies installed successfully!"
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "✅ .env file created"
else
    echo "ℹ️  .env file already exists"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "To start the development server, run:"
echo "  $PKG_MANAGER run dev"
echo ""
echo "The application will open at http://localhost:3000"
echo ""

