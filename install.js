const fs = require("fs");
const https = require("https");
const os = require("os");
const path = require("path");

const platform = os.platform();
const binDir = path.join(__dirname, "bin");

// Check if we already have a local binary (from local build)
const localBinPath = path.join(binDir, platform === "win32" ? "studio-mcp.exe" : "studio-mcp");
if (fs.existsSync(localBinPath)) {
  console.log("✅ Local binary already available:", localBinPath);
  return;
}

// Map platform names to binary names for downloads
const binName = platform === "darwin"
  ? "studio-mcp-macos"
  : platform === "linux"
  ? "studio-mcp-linux"
  : "studio-mcp-win.exe";

const url = `https://github.com/martinemde/studio-mcp/releases/latest/download/${binName}`;
const outPath = path.join(binDir, binName);

console.log(`Downloading ${binName} from GitHub releases...`);

fs.mkdirSync(binDir, { recursive: true });

https.get(url, res => {
  if (res.statusCode !== 200) {
    console.log(`No pre-built binary available (HTTP ${res.statusCode})`);
    console.log("Binary will be downloaded on first use or you can build from source.");
    return; // Exit gracefully without error
  }

  const file = fs.createWriteStream(outPath, { mode: 0o755 });
  res.pipe(file);

  file.on("finish", () => {
    file.close(() => {
      console.log("✅ Installed binary:", outPath);
    });
  });

  file.on("error", err => {
    fs.unlink(outPath, () => {}); // Delete incomplete file
    console.error("Error writing binary:", err.message);
    process.exit(1);
  });
}).on("error", err => {
  console.log("Binary not available for download:", err.message);
  console.log("Binary will be downloaded on first use or you can build from source.");
  // Exit gracefully without error
});
