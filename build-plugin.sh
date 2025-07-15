#!/bin/bash

# Build and package Grafana plugin for Helm chart deployment
set -e

echo "Building Grafana plugin..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "Error: Node.js is not installed. Please install Node.js 14+ first."
    exit 1
fi

# Check if Yarn is installed
if ! command -v yarn &> /dev/null; then
    echo "Error: Yarn is not installed. Please install Yarn first."
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 14 ]; then
    echo "Error: Node.js version 14+ is required. Current version: $(node -v)"
    exit 1
fi

echo "Node.js version: $(node -v)"
echo "Yarn version: $(yarn --version)"

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf dist/
rm -f *.zip

# Install dependencies
echo "Installing dependencies..."
yarn install --frozen-lockfile

# Build the plugin
echo "Building plugin..."
# Set Node.js compatibility for older webpack versions
export NODE_OPTIONS="--openssl-legacy-provider"
yarn build

# Check if build was successful
if [ ! -d "dist" ]; then
    echo "Error: Build failed - dist directory not found"
    exit 1
fi

# Get plugin metadata
echo "Extracting plugin metadata..."
if command -v jq &> /dev/null; then
    PLUGIN_ID=$(cat dist/plugin.json | jq -r .id)
    PLUGIN_VERSION=$(cat dist/plugin.json | jq -r .info.version)
    PLUGIN_TYPE=$(cat dist/plugin.json | jq -r .type)
else
    echo "Warning: jq not found, using default values"
    PLUGIN_ID="macropower-analytics-panel"
    PLUGIN_VERSION="2.1.0"
    PLUGIN_TYPE="panel"
fi

echo "Plugin ID: $PLUGIN_ID"
echo "Plugin Version: $PLUGIN_VERSION"
echo "Plugin Type: $PLUGIN_TYPE"

# Create zip file
ZIP_NAME="${PLUGIN_ID}-${PLUGIN_VERSION}.zip"
echo "Creating zip file: $ZIP_NAME"

# Create temporary directory with plugin name
TEMP_DIR=$(mktemp -d)
cp -r dist "$TEMP_DIR/$PLUGIN_ID"

# Create zip file
cd "$TEMP_DIR"
zip -r "$ZIP_NAME" "$PLUGIN_ID" -q
cd - > /dev/null

# Move zip file to current directory
mv "$TEMP_DIR/$ZIP_NAME" .

# Clean up
rm -rf "$TEMP_DIR"

# Create checksum
if command -v md5sum &> /dev/null; then
    md5sum "$ZIP_NAME" > "${ZIP_NAME}.md5"
    echo "MD5 checksum: $(cat ${ZIP_NAME}.md5)"
elif command -v md5 &> /dev/null; then
    md5 "$ZIP_NAME" > "${ZIP_NAME}.md5"
    echo "MD5 checksum: $(cat ${ZIP_NAME}.md5)"
fi

echo ""
echo "✅ Plugin built successfully!"
echo "📦 Zip file: $ZIP_NAME"
echo "📁 Size: $(du -h "$ZIP_NAME" | cut -f1)"
echo ""
echo "To use with Grafana Helm chart:"
echo "1. Copy $ZIP_NAME to your Helm chart's plugins directory"
echo "2. Update your values.yaml to include this plugin"
echo "3. Deploy with helm upgrade/install"
echo ""
echo "Example Helm values.yaml entry:"
echo "grafana:"
echo "  plugins:"
echo "    - name: $PLUGIN_ID"
echo "      version: $PLUGIN_VERSION"
echo "      source: $ZIP_NAME" 