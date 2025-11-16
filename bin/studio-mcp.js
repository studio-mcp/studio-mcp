#!/usr/bin/env node

// Simple alias wrapper for @studio-mcp/studio
import { createRequire } from "module";
const require = createRequire(import.meta.url);

// Re-export the actual CLI
const studioCli = require.resolve("@studio-mcp/studio/bin/studio-mcp.js");
await import(studioCli);
