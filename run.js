#!/usr/bin/env node
const path = require("path");
const { spawnSync } = require("child_process");
const fs = require("fs");
const os = require("os");

const platform = os.platform();
const arch = os.arch();
const binDir = path.join(__dirname, "bin");

// Normalize platform names to match our binary naming convention
const platformMap = {
  'darwin': 'macos',
  'linux': 'linux',
  'win32': 'win'
};

// For local development, use the simple binary name
// For distributed packages, use platform-specific names
const localBinPath = path.join(binDir, platform === "win32" ? "studio-mcp.exe" : "studio-mcp");
const normalizedPlatform = platformMap[platform] || platform;
const distributedBinName = platform === "win32" ? `studio-mcp-win.exe` : `studio-mcp-${normalizedPlatform}`;
const distributedBinPath = path.join(binDir, distributedBinName);

// Use local binary if it exists, otherwise use distributed binary path
const binPath = fs.existsSync(localBinPath) ? localBinPath : distributedBinPath;

// Check if binary exists, if not try to download it
if (!fs.existsSync(binPath)) {
  console.log("Binary not found, attempting to download...");
  const { execSync } = require("child_process");
  try {
    execSync("node " + path.join(__dirname, "install.js"), { stdio: "inherit" });
  } catch (error) {
    console.error("Failed to download binary. Please install from source using 'npm run build'");
    process.exit(1);
  }

  // Check again after download attempt
  if (!fs.existsSync(binPath)) {
    console.error("Binary still not available after download attempt.");
    console.error("Please build from source using 'npm run build:pkg' or check releases.");
    process.exit(1);
  }
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: "inherit" });
process.exit(result.status || 0);
