#!/bin/bash

set -e

echo "🔧 Setting up pre-commit hooks..."

if ! command -v pre-commit &> /dev/null; then
    echo "❌ pre-commit is not installed."
    echo ""
    echo "Please install pre-commit:"
    echo "  pip install pre-commit"
    echo "  or"
    echo "  brew install pre-commit"
    echo ""
    exit 1
fi

echo "✅ pre-commit is installed"

echo "📦 Installing pre-commit hooks..."
pre-commit install

echo "✅ Pre-commit hooks installed successfully"
echo ""
echo "To run hooks manually:"
echo "  pre-commit run --all-files"
echo ""
echo "To update hooks:"
echo "  pre-commit autoupdate"
