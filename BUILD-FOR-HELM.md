# Building Plugin for Grafana Helm Chart

This guide shows you how to build the macropower-analytics-panel plugin and create a zip file that can be used with Grafana Helm charts.

## Quick Build (Recommended)

Use the automated build script:

```bash
./build-plugin.sh
```

This script will:
- Check prerequisites (Node.js 14+, Yarn)
- Install dependencies
- Build the plugin
- Create a zip file with proper naming
- Generate MD5 checksum

## Manual Build Process

If you prefer to build manually:

### 1. Prerequisites
- Node.js 14 or higher
- Yarn package manager

### 2. Install Dependencies
```bash
yarn install --frozen-lockfile
```

### 3. Build the Plugin
```bash
yarn build
```

### 4. Create Zip File
```bash
# Get plugin metadata
PLUGIN_ID=$(cat dist/plugin.json | jq -r .id)
PLUGIN_VERSION=$(cat dist/plugin.json | jq -r .info.version)

# Create zip file
mv dist "$PLUGIN_ID"
zip -r "${PLUGIN_ID}-${PLUGIN_VERSION}.zip" "$PLUGIN_ID"
mv "$PLUGIN_ID" dist
```

## Using with Grafana Helm Chart

### Option 1: Local Plugin File

1. Copy the zip file to your Helm chart's plugins directory
2. Update your `values.yaml`:

```yaml
grafana:
  plugins:
    - name: macropower-analytics-panel
      version: 2.1.0
      source: macropower-analytics-panel-2.1.0.zip
```

### Option 2: URL-based Installation

If you host the zip file on a web server:

```yaml
grafana:
  plugins:
    - name: macropower-analytics-panel
      version: 2.1.0
      source: https://your-server.com/plugins/macropower-analytics-panel-2.1.0.zip
```

### Option 3: GitHub Release

If you create a GitHub release with the zip file:

```yaml
grafana:
  plugins:
    - name: macropower-analytics-panel
      version: 2.1.0
      source: https://github.com/your-org/macropower-analytics-panel/releases/download/v2.1.0/macropower-analytics-panel-2.1.0.zip
```

## Plugin Structure

The zip file contains:
```
macropower-analytics-panel/
├── plugin.json          # Plugin metadata
├── module.js            # Compiled plugin code
├── module.css           # Compiled styles
├── README.md            # Plugin documentation
└── ... (other assets)
```

## Troubleshooting

### Build Issues
- Ensure Node.js version is 14+
- Use `yarn install --frozen-lockfile` for consistent builds
- Check console output for specific error messages

### Helm Chart Issues
- Verify the plugin name matches exactly
- Ensure the zip file is accessible from the Kubernetes cluster
- Check Grafana logs for plugin loading errors

### Plugin Loading Issues
- Verify the plugin is enabled in Grafana
- Check browser console for JavaScript errors
- Ensure the plugin version is compatible with your Grafana version 