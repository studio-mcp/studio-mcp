import { spawn, ChildProcess } from 'child_process';
import { Studio } from '../src/studio';

// Helper function to send MCP requests to the server
function sendMcpRequest(
  commandArgs: string[],
  request: object,
  timeout: number = 5000
): Promise<any> {
  return new Promise((resolve, reject) => {
    const child = spawn('node', ['dist/cli.js', ...commandArgs], {
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let output = '';
    let errorOutput = '';
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

      // Try to parse JSON response(s)
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

    child.stderr?.on('data', (data) => {
      errorOutput += data.toString();
    });

    child.on('error', (error) => {
      if (!hasResolved) {
        hasResolved = true;
        clearTimeout(timeoutId);
        reject(new Error(`Server failed to start: ${error.message}`));
      }
    });

    child.on('exit', (code) => {
      if (!hasResolved) {
        hasResolved = true;
        clearTimeout(timeoutId);
        if (code !== 0) {
          reject(new Error(`Server exited with code ${code}: ${errorOutput}`));
        } else {
          reject(new Error('Server exited without sending response'));
        }
      }
    });

    // Send the request
    const requestJson = JSON.stringify(request);
    child.stdin?.write(requestJson + '\n');
    child.stdin?.end();
  });
}

describe('Studio MCP Server Integration', () => {
  beforeAll(async () => {
    // Make sure the project is built
    const { execSync } = require('child_process');
    try {
      execSync('npm run build', { stdio: 'inherit' });
    } catch (error) {
      console.error('Failed to build project:', error);
      throw error;
    }
  });

  describe('Basic MCP Protocol', () => {
    it('responds to ping requests', async () => {
      const request = {
        jsonrpc: '2.0',
        id: '1',
        method: 'ping'
      };

      const response = await sendMcpRequest(['echo', 'hello'], request);

      expect(response.jsonrpc).toBe('2.0');
      expect(response.id).toBe('1');
      expect(response.result).toEqual({});
    });

    it('responds to initialize requests', async () => {
      const request = {
        jsonrpc: '2.0',
        id: '1',
        method: 'initialize',
        params: {
          protocolVersion: '2024-11-05',
          capabilities: {},
          clientInfo: {
            name: 'test-client',
            version: '1.0.0'
          }
        }
      };

      const response = await sendMcpRequest(['echo', 'hello'], request);

      expect(response.jsonrpc).toBe('2.0');
      expect(response.id).toBe('1');
      expect(response.result).toHaveProperty('protocolVersion');
      expect(response.result).toHaveProperty('capabilities');
      expect(response.result).toHaveProperty('serverInfo');
      expect(response.result.serverInfo.name).toBe('studio-mcp');
    });
  });

  describe('Tools functionality', () => {
    describe('with simple echo command', () => {
      it('lists available tools', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '2',
          method: 'tools/list'
        };

        const response = await sendMcpRequest(['echo', 'hello', '[args...]'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('2');
        expect(response.result.tools).toBeInstanceOf(Array);
        expect(response.result.tools).toHaveLength(1);

        const tool = response.result.tools[0];
        expect(tool.name).toBe('echo');
        expect(tool.description).toBe('Run the shell command `echo hello [args...]`');
        expect(tool.inputSchema.type).toBe('object');
        expect(tool.inputSchema.properties).toHaveProperty('args');
      });

      it('executes simple echo command', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '3',
          method: 'tools/call',
          params: {
            name: 'echo',
            arguments: {
              args: ['hello', 'world']
            }
          }
        };

        const response = await sendMcpRequest(['echo', 'hello', '[args...]'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('3');
        expect(response.result.content).toBeInstanceOf(Array);
        expect(response.result.content[0].type).toBe('text');
        expect(response.result.content[0].text).toBe('hello hello world');
        expect(response.result.isError).toBe(false);
      });
    });

    describe('with blueprinted echo command', () => {
      it('lists blueprinted tool with proper schema', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '4',
          method: 'tools/list'
        };

        const response = await sendMcpRequest(['echo', '{{text#the text to echo}}'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('4');

        const tool = response.result.tools[0];
        expect(tool.name).toBe('echo');
        expect(tool.description).toBe('Run the shell command `echo {{text}}`');

        const schema = tool.inputSchema;
        expect(schema.properties).toHaveProperty('text');
        expect(schema.properties.text.type).toBe('string');
        expect(schema.properties.text.description).toBe('the text to echo');
        expect(schema.properties).not.toHaveProperty('args');
      });

      it('executes blueprinted command', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '5',
          method: 'tools/call',
          params: {
            name: 'echo',
            arguments: {
              text: 'Hello Blueprint!'
            }
          }
        };

        const response = await sendMcpRequest(['echo', '{{text#the text to echo}}'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('5');
        expect(response.result.content[0].type).toBe('text');
        expect(response.result.content[0].text).toBe('Hello Blueprint!');
        expect(response.result.isError).toBe(false);
      });

      it('executes blueprinted command with additional args', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '6',
          method: 'tools/call',
          params: {
            name: 'echo',
            arguments: {
              text: 'Hello',
              args: ['World', 'from', 'args']
            }
          }
        };

        const response = await sendMcpRequest(['echo', '{{text#the text to echo}}', '[args...]'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('6');
        expect(response.result.content[0].type).toBe('text');
        expect(response.result.content[0].text).toBe('Hello World from args');
        expect(response.result.isError).toBe(false);
      });

      it('handles blueprint arguments with spaces', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '7',
          method: 'tools/call',
          params: {
            name: 'echo',
            arguments: {
              text: 'Hello World with spaces'
            }
          }
        };

        const response = await sendMcpRequest(['echo', '{{text#the text to echo}}'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('7');
        expect(response.result.content[0].type).toBe('text');
        expect(response.result.content[0].text).toBe('Hello World with spaces');
        expect(response.result.isError).toBe(false);
      });

      it('handles mixed blueprint with spaces', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '8',
          method: 'tools/call',
          params: {
            name: 'echo',
            arguments: {
              text: 'Hello World'
            }
          }
        };

        const response = await sendMcpRequest(['echo', 'simon says {{text#the text to echo}}'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('8');
        expect(response.result.content[0].type).toBe('text');
        expect(response.result.content[0].text).toBe('simon says Hello World');
        expect(response.result.isError).toBe(false);
      });

      it('handles blueprint definitions with spaces around hash', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '9',
          method: 'tools/call',
          params: {
            name: 'echo',
            arguments: {
              text: 'Hello World'
            }
          }
        };

        const response = await sendMcpRequest(['echo', '{{text # the text to echo}}'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('9');
        expect(response.result.content[0].type).toBe('text');
        expect(response.result.content[0].text).toBe('Hello World');
        expect(response.result.isError).toBe(false);
      });

      it('handles blueprints without descriptions', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '10',
          method: 'tools/list'
        };

        const response = await sendMcpRequest(['echo', '{{text}}'], request);

        const tool = response.result.tools[0];
        expect(tool.name).toBe('echo');
        expect(tool.description).toBe('Run the shell command `echo {{text}}`');

        const schema = tool.inputSchema;
        expect(schema.properties).toHaveProperty('text');
        expect(schema.properties.text.type).toBe('string');
        expect(schema.properties.text.description).toBeUndefined();
      });

      it('executes blueprints without descriptions', async () => {
        const request = {
          jsonrpc: '2.0',
          id: '11',
          method: 'tools/call',
          params: {
            name: 'echo',
            arguments: {
              text: 'Hello Blueprint!'
            }
          }
        };

        const response = await sendMcpRequest(['echo', '{{text}}'], request);

        expect(response.jsonrpc).toBe('2.0');
        expect(response.id).toBe('11');
        expect(response.result.content[0].type).toBe('text');
        expect(response.result.content[0].text).toBe('Hello Blueprint!');
        expect(response.result.isError).toBe(false);
      });
    });
  });

  describe('Error handling', () => {
    it('handles command errors gracefully', async () => {
      const request = {
        jsonrpc: '2.0',
        id: '12',
        method: 'tools/call',
        params: {
          name: 'false',
          arguments: {}
        }
      };

      const response = await sendMcpRequest(['false'], request);

      expect(response.jsonrpc).toBe('2.0');
      expect(response.id).toBe('12');
      expect(response.result.isError).toBe(true);
      expect(response.result.content).toBeInstanceOf(Array);
    });

    it('handles nonexistent tools', async () => {
      const request = {
        jsonrpc: '2.0',
        id: '13',
        method: 'tools/call',
        params: {
          name: 'nonexistent',
          arguments: {}
        }
      };

      const response = await sendMcpRequest(['echo', 'hello'], request);

      expect(response.jsonrpc).toBe('2.0');
      expect(response.id).toBe('13');
      expect(response).toHaveProperty('error');
    });
  });

  describe('Complex commands', () => {
    it('works with git commands', async () => {
      const request = {
        jsonrpc: '2.0',
        id: '14',
        method: 'tools/list'
      };

      const response = await sendMcpRequest(['git', 'status', '[args...]'], request);

      const tool = response.result.tools[0];
      expect(tool.name).toBe('git');
      expect(tool.description).toBe('Run the shell command `git status [args...]`');
    });

    it('works with multiple blueprint arguments', async () => {
      const request = {
        jsonrpc: '2.0',
        id: '15',
        method: 'tools/list'
      };

      const response = await sendMcpRequest([
        'rails', 'generate',
        '{{generator#Rails generator name}}',
        '{{name#Resource name}}'
      ], request);

      const tool = response.result.tools[0];
      expect(tool.name).toBe('rails');
      expect(tool.description).toBe('Run the shell command `rails generate {{generator}} {{name}}`');

      const schema = tool.inputSchema;
      expect(schema.properties).toHaveProperty('generator');
      expect(schema.properties).toHaveProperty('name');
      expect(schema.properties).not.toHaveProperty('args');
    });
  });
});
