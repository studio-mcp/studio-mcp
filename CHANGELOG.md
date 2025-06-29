# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2024-12-19

### Added
- Initial TypeScript port of `studio` from Ruby
- MCP server implementation using @modelcontextprotocol/sdk
- Blueprint templating
- Support for required `{{name}}` and optional `[name]` string arguments
- Support for array arguments `[name...]`
- CLI interface compatible with Claude Desktop, Cursor, and other MCP clients
