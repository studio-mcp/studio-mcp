export interface FieldToken {
  type: 'field';
  name: string;
  description?: string;
  content: string;
  required: boolean;
  fieldType: 'string' | 'array';
}

export interface TextToken {
  type: 'text';
  content: string;
}

export type Token = FieldToken | TextToken;

export interface PropertySchema {
  type: string;
  description?: string;
  items?: { type: string };
}

export interface InputSchema {
  type: 'object';
  properties: Record<string, PropertySchema>;
  required?: string[];
}

export class Blueprint {
  public readonly argv: string[];
  public readonly baseCommand: string;
  public readonly compiledArgs: Token[][];
  private properties: Record<string, PropertySchema & { required?: boolean }> = {};

  constructor(argv: string[]) {
    this.argv = [...argv];
    this.baseCommand = this.argv.shift() || '';
    this.compiledArgs = this.compileTemplateArgs(this.argv);
  }

  static argv(argv: string[]): Blueprint {
    return new Blueprint(argv);
  }

  /**
   * Build the command with provided arguments
   */
  buildCommandArgs(args: Record<string, any> = {}): string[] {
    return this.buildCommand(args);
  }

  /**
   * Get the tool name for MCP
   */
  get toolName(): string {
    return this.baseCommand.replace(/[^a-zA-Z0-9_]/g, '_');
  }

  /**
   * Get the tool description for MCP
   */
  get toolDescription(): string {
    return `Run the shell command \`${this.formatCommand()}\``;
  }

  /**
   * Build display format directly from compiled args
   */
  formatCommand(): string {
    const displayParts = [this.baseCommand];

    for (const compiledArg of this.compiledArgs) {
      let processedArg = '';

      for (const part of compiledArg) {
        if (part.type === 'text') {
          processedArg += part.content;
        } else if (part.type === 'field') {
          const normalizedName = this.normalizeFieldName(part.name);
          if (part.fieldType === 'array') {
            processedArg = `[${normalizedName}...]`;
            break;
          } else if (part.required) {
            processedArg += `{{${normalizedName}}}`;
          } else {
            processedArg += `[${normalizedName}]`;
          }
        }
      }

      displayParts.push(processedArg);
    }

    return displayParts
      .map(part => part.includes(' ') ? `"${part}"` : part)
      .join(' ');
  }

  /**
   * Get the input schema for MCP tool
   */
  get inputSchema(): InputSchema {
    // Build properties without the internal required field
    const cleanProperties: Record<string, PropertySchema> = {};

    for (const [key, prop] of Object.entries(this.properties)) {
      const { required, ...cleanProp } = prop;
      cleanProperties[key] = cleanProp;
    }

    const schema: InputSchema = {
      type: 'object',
      properties: cleanProperties
    };

    const requiredFields = Object.entries(this.properties)
      .filter(([_, prop]) => prop.required)
      .map(([key, _]) => key);

    if (requiredFields.length > 0) {
      schema.required = requiredFields;
    }

    return schema;
  }

  /**
   * Lexer: tokenizes a single shell word into tokens
   */
  private *lex(word: string): Generator<Token> {
    // Split on template boundaries while capturing the delimiters
    const parts = word.split(/(\{\{[^}]*\}\}|\[[^\]]*\])/);

    for (const part of parts) {
      if (!part) continue;

      if (part.match(/^\{\{.*\}\}$/)) {
        // Extract field content without the braces
        const fieldContent = part.slice(2, -2);
        const [name, description] = fieldContent.split('#', 2);
        yield {
          type: 'field',
          name: name.trim(),
          description: description?.trim(),
          content: part,
          required: true,
          fieldType: 'string'
        };
      } else if (part.match(/^\[.*\]$/) && part === word.trim()) {
        // Only accept arrays that are the entire shell word
        const fieldContent = part.slice(1, -1);
        const [nameWithEllipsis, description] = fieldContent.split('#', 2);
        const name = nameWithEllipsis.trim();

        if (name.endsWith('...')) {
          yield {
            type: 'field',
            name: name.slice(0, -3),
            description: description?.trim(),
            content: part,
            required: true,
            fieldType: 'array'
          };
        } else {
          yield {
            type: 'field',
            name: name,
            description: description?.trim(),
            content: part,
            required: false,
            fieldType: 'string'
          };
        }
      } else {
        yield {
          type: 'text',
          content: part
        };
      }
    }
  }

  private compileTemplateArgs(args: string[]): Token[][] {
    return args.map(arg => this.compileTemplateArg(arg));
  }

  private compileTemplateArg(arg: string): Token[] {
    const parts: Token[] = [];

    for (const token of this.lex(arg)) {
      if (token.type === 'field') {
        this.addProperty(token);
      }
      parts.push(token);
    }

    return parts;
  }

  /**
   * Normalize field name by converting dashes to underscores
   */
  private normalizeFieldName(name: string): string {
    return name.replace(/-/g, '_');
  }

  /**
   * Update properties, allowing that a variable may be used twice, not always with the description
   */
  private addProperty(token: FieldToken): void {
    const { name, description, fieldType, required } = token;
    const normalizedName = this.normalizeFieldName(name);

    if (!this.properties[normalizedName]) {
      this.properties[normalizedName] = { type: fieldType, required };
    }

    this.properties[normalizedName].type = fieldType;
    this.properties[normalizedName].required = required; // Keep this for internal tracking

    if (fieldType === 'array') {
      this.properties[normalizedName].items = { type: 'string' };
      this.properties[normalizedName].description = description?.trim() || 'Additional command line arguments';
    } else if (description) {
      this.properties[normalizedName].description = description.trim();
    }
  }

  private buildCommand(args: Record<string, any>): string[] {
    const commandParts = [this.baseCommand];

    // Process compiled arguments
    for (const compiledArg of this.compiledArgs) {
      let processedArg = '';
      let skipArg = false;

      for (const part of compiledArg) {
        if (part.type === 'text') {
          processedArg += part.content;
        } else if (part.type === 'field') {
          const normalizedName = this.normalizeFieldName(part.name);
          if (part.fieldType === 'array') {
            // For array arguments, we expect them to be passed as the whole argument
            const value = args[normalizedName] || [];
            const arrayValue = Array.isArray(value) ? value : [value];
            if (arrayValue.length > 0) {
              commandParts.push(...arrayValue);
            }
            processedArg = ''; // Skip adding this as a single argument
            break;
          } else {
            const value = args[normalizedName];
            if (value === null || value === undefined || value === '') {
              if (part.required) {
                processedArg += '';
              } else {
                // Skip the entire argument if it's optional and not provided
                skipArg = true;
                break;
              }
            } else {
              processedArg += String(value);
            }
          }
        }
      }

      if (!skipArg && processedArg !== '') {
        commandParts.push(processedArg);
      }
    }

    return commandParts;
  }
}
