import { spawn } from 'child_process';

/**
 * Test the exact examples from the README that were requested
 *
 * Example usage:
 * ```bash
 * $ npx -y studio-mcp echo "{{message}}"
 * ```
 *
 * Then interact with JSON-RPC:
 * - ping: {"jsonrpc":"2.0","id":"1","method":"ping"}
 * - initialize: {"jsonrpc": "2.0","id": "2","method": "initialize","params": {"protocolVersion": "2024-11-05","capabilities": {},"clientInfo": {"name": "test-client","version": "1.0.0"}}}
 * - tools/list: {"jsonrpc":"2.0","id":"3","method":"tools/list"}
 * - tools/call: {"jsonrpc": "2.0","id": "4","method": "tools/call","params": {"name": "echo","arguments": {"message": "Hello world"}}}
 */

function sendJsonRpcRequest(
  commandArgs: string[],
  request: object,
  timeout: number = 5000
): Promise<any> {
  return new Promise((resolve, reject) => {
    const child = spawn('node', ['dist/cli.js', ...commandArgs], {
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let output = '';
    let hasResolved = false;

    const timeoutId = setTimeout(() => {
      if (!hasResolved) {
        hasResolved = true;
        child.kill();
        reject(new Error(`Request timed out after ${timeout}ms`));
      }
    }, timeout);

    child.stdout?.on('data', (data) => {
      output += data.toString();

      const lines = output.split('\n').filter(line => line.trim());
      if (lines.length > 0) {
        try {
          const response = JSON.parse(lines[lines.length - 1]);
          if (!hasResolved) {
            hasResolved = true;
            clearTimeout(timeoutId);
            child.kill();
            resolve(response);
          }
        } catch (e) {
          // Not valid JSON yet, continue waiting
        }
      }
    });

    child.on('error', (error) => {
      if (!hasResolved) {
        hasResolved = true;
        clearTimeout(timeoutId);
        reject(error);
      }
    });

    // Send the request
    const requestJson = JSON.stringify(request);
    child.stdin?.write(requestJson + '\n');
    child.stdin?.end();
  });
}

describe('README Examples - Exact Command Line Usage', () => {
  beforeAll(async () => {
    const { execSync } = require('child_process');
    try {
      execSync('npm run build', { stdio: 'inherit' });
    } catch (error) {
      console.error('Failed to build project:', error);
      throw error;
    }
  });

  describe('Command: npx -y studio-mcp echo "{{message}}"', () => {
    const commandArgs = ['echo', '{{message}}'];

    it('ping request', async () => {
      const request = {"jsonrpc":"2.0","id":"1","method":"ping"};
      const response = await sendJsonRpcRequest(commandArgs, request);

      expect(response).toEqual({
        "result": {},
        "jsonrpc": "2.0",
        "id": "1"
      });
    });

    it('initialize request', async () => {
      const request = {
        "jsonrpc": "2.0",
        "id": "2",
        "method": "initialize",
        "params": {
          "protocolVersion": "2024-11-05",
          "capabilities": {},
          "clientInfo": {
            "name": "test-client",
            "version": "1.0.0"
          }
        }
      };

      const response = await sendJsonRpcRequest(commandArgs, request);

      expect(response.jsonrpc).toBe("2.0");
      expect(response.id).toBe("2");
      expect(response.result).toHaveProperty('protocolVersion');
      expect(response.result).toHaveProperty('capabilities');
      expect(response.result).toHaveProperty('serverInfo');
      expect(response.result.serverInfo.name).toBe('studio-mcp');
    });

    it('tools/list request', async () => {
      const request = {"jsonrpc":"2.0","id":"3","method":"tools/list"};
      const response = await sendJsonRpcRequest(commandArgs, request);

      expect(response.jsonrpc).toBe("2.0");
      expect(response.id).toBe("3");
      expect(response.result.tools).toHaveLength(1);
      expect(response.result.tools[0].name).toBe("echo");
      expect(response.result.tools[0].inputSchema.properties.message).toEqual({
        type: "string"
      });
    });

    it('tools/call request', async () => {
      const request = {
        "jsonrpc": "2.0",
        "id": "4",
        "method": "tools/call",
        "params": {
          "name": "echo",
          "arguments": {
            "message": "Hello world"
          }
        }
      };

      const response = await sendJsonRpcRequest(commandArgs, request);

      expect(response.jsonrpc).toBe("2.0");
      expect(response.id).toBe("4");
      expect(response.result.content[0].type).toBe("text");
      expect(response.result.content[0].text).toBe("Hello world");
      expect(response.result.isError).toBe(false);
    });
  });
});
