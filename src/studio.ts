import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import { Blueprint } from './blueprint';
import { Tool } from './tool';

export class Studio {
  private server: Server;
  private blueprint: Blueprint;
  private debugMode: boolean = false;

  constructor(argv: string[]) {
        if (argv.length === 0) {
      throw new Error('Usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
    }

    const { flags, commandArgs } = this.parseArguments(argv);

    if (commandArgs.length === 0) {
      throw new Error('Usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
    }

    this.debugMode = flags.includes('--debug');
    this.blueprint = new Blueprint(commandArgs);

    // Set debug mode on Tool class
    Tool.setDebugMode(this.debugMode);

    this.server = new Server(
      {
        name: 'studio-mcp',
        version: '1.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.setupHandlers();
  }

  private parseArguments(argv: string[]): { flags: string[], commandArgs: string[] } {
    const flags: string[] = [];
    let commandStartIndex = 0;

    // Find the first non-flag argument (doesn't start with -)
    for (let i = 0; i < argv.length; i++) {
      if (argv[i].startsWith('-')) {
        flags.push(argv[i]);
        commandStartIndex = i + 1;
      } else {
        // Found the first non-flag argument, this is where the command starts
        commandStartIndex = i;
        break;
      }
    }

    const commandArgs = argv.slice(commandStartIndex);
    return { flags, commandArgs };
  }

  get isDebugMode(): boolean {
    return this.debugMode;
  }

  private setupHandlers(): void {
    // List available tools
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: [
          {
            name: this.blueprint.toolName,
            description: this.blueprint.toolDescription,
            inputSchema: this.blueprint.inputSchema,
          },
        ],
      };
    });

    // Handle tool calls
    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;

      if (name !== this.blueprint.toolName) {
        throw new Error(`Unknown tool: ${name}`);
      }

      const toolFunction = Tool.createToolFunction(this.blueprint);
      const result = await toolFunction(args || {});

      return result;
    });
  }

  /**
   * Start the MCP server
   */
  async serve(): Promise<void> {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
  }

  /**
   * Static method to create and start a server
   */
  static async serve(argv: string[]): Promise<void> {
    try {
      const studio = new Studio(argv);
      await studio.serve();
    } catch (error) {
      console.error(`Studio error: ${error instanceof Error ? error.message : String(error)}`);
      process.exit(1);
    }
  }
}
