import { spawn } from 'child_process';
import { Blueprint } from './blueprint';

export interface ExecuteResult {
  output: string;
  success: boolean;
}

export class Tool {
  private static debugEnabled: boolean = false;

  static setDebugMode(enabled: boolean): void {
    this.debugEnabled = enabled;
  }

  private static isDebugMode(): boolean {
    return process.env.NODE_ENV !== 'test' && this.debugEnabled;
  }

  private static debug(...args: any[]): void {
    if (this.isDebugMode()) {
      console.error(...args);
    }
  }

  /**
   * Execute the command with provided arguments
   */
    static async execute(...command: string[]): Promise<ExecuteResult> {
    // Log to stderr for MCP debugging (won't interfere with stdout protocol)
    Tool.debug(`[Studio MCP] Executing command: ${command.join(' ')}`);

    try {
      const [cmd, ...args] = command;

      if (!cmd || cmd.trim() === '') {
        const errorMsg = 'Studio error: Empty command provided';
        Tool.debug(`[Studio MCP] Error: ${errorMsg}`);
        return {
          output: errorMsg,
          success: false
        };
      }

      return new Promise((resolve) => {
        const child = spawn(cmd, args, {
          stdio: ['pipe', 'pipe', 'pipe'],
          shell: false
        });

        let output = '';
        let errorOutput = '';

        child.stdout?.on('data', (data) => {
          const chunk = data.toString();
          output += chunk;
          Tool.debug(`[Studio MCP] stdout: ${chunk.trim()}`);
        });

        child.stderr?.on('data', (data) => {
          const chunk = data.toString();
          errorOutput += chunk;
          Tool.debug(`[Studio MCP] stderr: ${chunk.trim()}`);
        });

        child.on('close', (code) => {
          const fullOutput = output + (errorOutput ? '\n' + errorOutput : '');
          Tool.debug(`[Studio MCP] Command completed with exit code: ${code}`);
          Tool.debug(`[Studio MCP] Final output length: ${fullOutput.length} chars`);

          resolve({
            output: fullOutput.trim(),
            success: code === 0
          });
        });

        child.on('error', (error) => {
          const errorMsg = `Studio error: ${error.message}`;
          Tool.debug(`[Studio MCP] Spawn error: ${errorMsg}`);
          resolve({
            output: errorMsg,
            success: false
          });
        });
      });
    } catch (error) {
      const errorMsg = `Studio error: ${error instanceof Error ? error.message : String(error)}`;
      Tool.debug(`[Studio MCP] Catch error: ${errorMsg}`);
      return {
        output: errorMsg,
        success: false
      };
    }
  }

  /**
   * Create a tool function for the given blueprint
   */
  static createToolFunction(blueprint: Blueprint) {
    return async (args: Record<string, any> = {}) => {
      Tool.debug(`[Studio MCP] Tool called with args:`, JSON.stringify(args, null, 2));

      const fullCommand = blueprint.buildCommandArgs(args);
      Tool.debug(`[Studio MCP] Built command: ${fullCommand.join(' ')}`);

      const { output, success } = await Tool.execute(...fullCommand);

      const result = {
        content: [{ type: 'text', text: output }],
        isError: !success
      };

      Tool.debug(`[Studio MCP] Tool result - success: ${success}, output length: ${output.length}`);

      return result;
    };
  }
}
