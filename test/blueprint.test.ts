import { Blueprint } from '../src/blueprint';
import { Tool } from '../src/tool';

describe('Blueprint', () => {
  describe('argv parsing', () => {
    it('parses simple command with explicit args', () => {
      const blueprint = Blueprint.argv(['git', 'status', '[args...]']);
      expect(blueprint.baseCommand).toBe('git');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          args: {
            type: 'array',
            items: { type: 'string' },
            description: 'Additional command line arguments'
          }
        },
        required: ['args']
      });
    });

    it('parses simple command without args', () => {
      const blueprint = Blueprint.argv(['git', 'status']);
      expect(blueprint.baseCommand).toBe('git');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {}
      });
    });

    it('parses blueprinted command', () => {
      const blueprint = Blueprint.argv(['curl', 'https://en.m.wikipedia.org/wiki/{{page#A valid wikipedia page}}']);
      expect(blueprint.baseCommand).toBe('curl');
      expect(blueprint.toolName).toBe('curl');
      expect(blueprint.toolDescription).toBe('Run the shell command `curl https://en.m.wikipedia.org/wiki/{{page}}`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          page: { type: 'string', description: 'A valid wikipedia page' }
        },
        required: ['page']
      });
    });

    it('parses blueprinted command with spaces between tokens', () => {
      const blueprint = Blueprint.argv(['curl', 'https://en.m.wikipedia.org/wiki/{{page # A valid wikipedia page}}']);
      expect(blueprint.baseCommand).toBe('curl');
      expect(blueprint.toolName).toBe('curl');
      expect(blueprint.toolDescription).toBe('Run the shell command `curl https://en.m.wikipedia.org/wiki/{{page}}`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          page: { type: 'string', description: 'A valid wikipedia page' }
        },
        required: ['page']
      });
    });

    it('parses blueprinted command without description', () => {
      const blueprint = Blueprint.argv(['echo', '{{text}}']);
      expect(blueprint.baseCommand).toBe('echo');
      expect(blueprint.toolName).toBe('echo');
      expect(blueprint.toolDescription).toBe('Run the shell command `echo {{text}}`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          text: { type: 'string' }
        },
        required: ['text']
      });
    });

    it('parses mixed blueprints with required and optional arguments', () => {
      const blueprint = Blueprint.argv(['command', '{{arg1#Custom description}}', '[arg2#Optional argument]']);
      expect(blueprint.baseCommand).toBe('command');
      expect(blueprint.toolName).toBe('command');
      expect(blueprint.toolDescription).toBe('Run the shell command `command {{arg1}} [arg2]`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          arg1: { type: 'string', description: 'Custom description' },
          arg2: { type: 'string', description: 'Optional argument' }
        },
        required: ['arg1']
      });
    });

    it('prioritizes explicit description over default', () => {
      const blueprint = Blueprint.argv(['echo', '{{text#Explicit description}}', '{{text}}']);
      expect(blueprint.baseCommand).toBe('echo');
      expect(blueprint.toolName).toBe('echo');
      expect(blueprint.toolDescription).toBe('Run the shell command `echo {{text}} {{text}}`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          text: { type: 'string', description: 'Explicit description' }
        },
        required: ['text']
      });
    });

    it('parses array arguments', () => {
      const blueprint = Blueprint.argv(['echo', '[files...#Files to echo]']);
      expect(blueprint.baseCommand).toBe('echo');
      expect(blueprint.toolName).toBe('echo');
      expect(blueprint.toolDescription).toBe('Run the shell command `echo [files...]`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          files: {
            type: 'array',
            items: { type: 'string' },
            description: 'Files to echo'
          }
        },
        required: ['files']
      });
    });

    it('parses array arguments without description', () => {
      const blueprint = Blueprint.argv(['ls', '[paths...]']);
      expect(blueprint.baseCommand).toBe('ls');
      expect(blueprint.toolName).toBe('ls');
      expect(blueprint.toolDescription).toBe('Run the shell command `ls [paths...]`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          paths: {
            type: 'array',
            items: { type: 'string' },
            description: 'Additional command line arguments'
          }
        },
        required: ['paths']
      });
    });

    it('parses [optional] (without ellipsis) as optional string field', () => {
      const blueprint = Blueprint.argv(['echo', '[optional]']);
      expect(blueprint.baseCommand).toBe('echo');
      expect(blueprint.toolName).toBe('echo');
      expect(blueprint.toolDescription).toBe('Run the shell command `echo [optional]`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          'optional': { type: 'string' }
        }
      });
    });

    it('converts dashes to understores in argument names', () => {
      const blueprint = Blueprint.argv(['echo', '[has-dashes]']);
      expect(blueprint.baseCommand).toBe('echo');
      expect(blueprint.toolName).toBe('echo');
      expect(blueprint.toolDescription).toBe('Run the shell command `echo [has_dashes]`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          'has_dashes': { type: 'string' }
        }
      });
    });

    it('parses mixed string and array arguments', () => {
      const blueprint = Blueprint.argv(['command', '{{flag#Command flag}}', '[files...#Files to process]']);
      expect(blueprint.baseCommand).toBe('command');
      expect(blueprint.toolName).toBe('command');
      expect(blueprint.toolDescription).toBe('Run the shell command `command {{flag}} [files...]`');
      expect(blueprint.inputSchema).toEqual({
        type: 'object',
        properties: {
          flag: { type: 'string', description: 'Command flag' },
          files: {
            type: 'array',
            items: { type: 'string' },
            description: 'Files to process'
          }
        },
        required: ['flag', 'files']
      });
    });
  });

  describe('toolName', () => {
    it('converts command to valid tool name', () => {
      const blueprint = Blueprint.argv(['git-flow']);
      expect(blueprint.toolName).toBe('git_flow');
    });
  });

  describe('toolDescription', () => {
    it('generates description for simple command without args', () => {
      const blueprint = Blueprint.argv(['git']);
      expect(blueprint.toolDescription).toBe('Run the shell command `git`');
    });

    it('generates description for simple command with explicit args', () => {
      const blueprint = Blueprint.argv(['git', '[args...]']);
      expect(blueprint.toolDescription).toBe('Run the shell command `git [args...]`');
    });

    it('generates description for blueprinted command', () => {
      const blueprint = Blueprint.argv(['rails', 'generate', '{{generator#A rails generator}}']);
      expect(blueprint.toolDescription).toBe('Run the shell command `rails generate {{generator}}`');
    });
  });

  describe('command execution', () => {
    it('executes simple command', async () => {
      const blueprint = Blueprint.argv(['echo', 'hello']);
      const fullCommand = blueprint.buildCommandArgs();
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('hello');
    });

    it('handles command errors', async () => {
      const blueprint = Blueprint.argv(['false']);
      const fullCommand = blueprint.buildCommandArgs();
      const { output, success } = await Tool.execute(...fullCommand);

      expect(output).toBe('');
      expect(success).toBe(false);
    });

    it('handles blueprint arguments with spaces', async () => {
      const blueprint = Blueprint.argv(['echo', '{{text#text to echo}}']);
      const fullCommand = blueprint.buildCommandArgs({ text: 'Hello World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('Hello World');
    });

    it('handles blueprint arguments without descriptions', async () => {
      const blueprint = Blueprint.argv(['echo', '{{text}}']);
      const fullCommand = blueprint.buildCommandArgs({ text: 'Hello World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('Hello World');
    });

    it('handles mixed blueprint with and without descriptions', async () => {
      const blueprint = Blueprint.argv(['echo', '{{greeting#The greeting}}', '{{name}}']);
      const fullCommand = blueprint.buildCommandArgs({ greeting: 'Hello', name: 'World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('Hello World');
    });

    it('handles blueprint arguments with spaces in mixed content', async () => {
      const blueprint = Blueprint.argv(['echo', 'simon says {{text#text for simon to say}}']);
      const fullCommand = blueprint.buildCommandArgs({ text: 'Hello World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('simon says Hello World');
    });

    it('handles blueprint arguments with special shell characters', async () => {
      const blueprint = Blueprint.argv(['echo', '{{text#text to echo}}']);
      const fullCommand = blueprint.buildCommandArgs({ text: 'Hello & World; echo pwned' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('Hello & World; echo pwned');
    });

    it('handles multiple blueprint substitutions in one argument', async () => {
      const blueprint = Blueprint.argv(['echo', '{{greeting}} {{name}}!']);
      const fullCommand = blueprint.buildCommandArgs({ greeting: 'Hello', name: 'World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('Hello World!');
    });

    it('handles blueprint in the middle of argument', async () => {
      const blueprint = Blueprint.argv(['echo', '--message={{text#message content}}']);
      const fullCommand = blueprint.buildCommandArgs({ text: 'Hello World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('--message=Hello World');
    });

    it('handles blueprint with prefix and suffix', async () => {
      const blueprint = Blueprint.argv(['echo', 'prefix-{{text#middle part}}-suffix']);
      const fullCommand = blueprint.buildCommandArgs({ text: 'Hello World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('prefix-Hello World-suffix');
    });

    it('handles mixed blueprint and non-blueprint arguments', async () => {
      const blueprint = Blueprint.argv(['echo', 'static', '{{dynamic#dynamic content}}', 'more-static']);
      const fullCommand = blueprint.buildCommandArgs({ dynamic: 'Hello World' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('static Hello World more-static');
    });

    it('preserves shell safety with complex blueprint values', async () => {
      const blueprint = Blueprint.argv(['echo', 'Result: {{text#text content}}']);
      const fullCommand = blueprint.buildCommandArgs({ text: "$(echo 'dangerous'); echo 'safe'" });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe("Result: $(echo 'dangerous'); echo 'safe'");
    });

    it('handles array arguments', async () => {
      const blueprint = Blueprint.argv(['echo', '[files...#Files to echo]']);
      const fullCommand = blueprint.buildCommandArgs({ files: ['file1.txt', 'file2.txt'] });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('file1.txt file2.txt');
    });

    it('handles empty array arguments', async () => {
      const blueprint = Blueprint.argv(['echo', 'hello', '[files...]']);
      const fullCommand = blueprint.buildCommandArgs({ files: [] });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('hello');
    });

    it('handles mixed string and array arguments', async () => {
      const blueprint = Blueprint.argv(['echo', '{{prefix#Prefix text}}', '[files...#Files to list]']);
      const fullCommand = blueprint.buildCommandArgs({ prefix: 'Files:', files: ['a.txt', 'b.txt'] });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('Files: a.txt b.txt');
    });

    it('handles optional arguments when provided', async () => {
      const blueprint = Blueprint.argv(['echo', 'hello', '[name]']);
      const fullCommand = blueprint.buildCommandArgs({ name: 'world' });
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('hello world');
    });

    it('skips optional arguments when not provided', async () => {
      const blueprint = Blueprint.argv(['echo', 'hello', '[name]']);
      const fullCommand = blueprint.buildCommandArgs({});
      const { output, success } = await Tool.execute(...fullCommand);

      expect(success).toBe(true);
      expect(output).toBe('hello');
    });

    it('handles mixed required and optional arguments', async () => {
      const blueprint = Blueprint.argv(['echo', '{{required#Required text}}', '[optional#Optional text]']);

      // With both provided
      let fullCommand = blueprint.buildCommandArgs({ required: 'hello', optional: 'world' });
      let { output, success } = await Tool.execute(...fullCommand);
      expect(success).toBe(true);
      expect(output).toBe('hello world');

      // With only required provided
      fullCommand = blueprint.buildCommandArgs({ required: 'hello' });
      ({ output, success } = await Tool.execute(...fullCommand));
      expect(success).toBe(true);
      expect(output).toBe('hello');
    });
  });
});
