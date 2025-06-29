import { Tool } from '../src/tool';
import { Blueprint } from '../src/blueprint';

describe('Tool', () => {
  describe('execute', () => {
    it('executes simple command successfully', async () => {
      const { output, success } = await Tool.execute('echo', 'hello');
      expect(success).toBe(true);
      expect(output).toBe('hello');
    });

    it('handles command failure', async () => {
      const { output, success } = await Tool.execute('false');
      expect(success).toBe(false);
      expect(output).toBe('');
    });

    it('handles nonexistent command', async () => {
      const { output, success } = await Tool.execute('nonexistent-command-12345');
      expect(success).toBe(false);
      expect(output).toContain('Studio error:');
    });

    it('handles command with arguments', async () => {
      const { output, success } = await Tool.execute('echo', 'hello', 'world');
      expect(success).toBe(true);
      expect(output).toBe('hello world');
    });

    it('captures stderr output', async () => {
      const { output, success } = await Tool.execute('sh', '-c', 'echo "error" >&2');
      expect(success).toBe(true);
      expect(output).toContain('error');
    });

    it('handles empty command', async () => {
      const { output, success } = await Tool.execute('');
      expect(success).toBe(false);
      expect(output).toContain('Studio error:');
    });
  });

  describe('createToolFunction', () => {
    it('creates a tool function that executes commands', async () => {
      const blueprint = Blueprint.argv(['echo', '{{text}}']);
      const toolFunction = Tool.createToolFunction(blueprint);

      const result = await toolFunction({ text: 'hello world' });

      expect(result.content).toHaveLength(1);
      expect(result.content[0].type).toBe('text');
      expect(result.content[0].text).toBe('hello world');
      expect(result.isError).toBe(false);
    });

    it('handles tool function errors', async () => {
      const blueprint = Blueprint.argv(['false']);
      const toolFunction = Tool.createToolFunction(blueprint);

      const result = await toolFunction({});

      expect(result.content).toHaveLength(1);
      expect(result.content[0].type).toBe('text');
      expect(result.isError).toBe(true);
    });

    it('handles array arguments in tool function', async () => {
      const blueprint = Blueprint.argv(['echo', '[args...]']);
      const toolFunction = Tool.createToolFunction(blueprint);

      const result = await toolFunction({ args: ['hello', 'world', 'test'] });

      expect(result.content).toHaveLength(1);
      expect(result.content[0].type).toBe('text');
      expect(result.content[0].text).toBe('hello world test');
      expect(result.isError).toBe(false);
    });

    it('handles empty arguments in tool function', async () => {
      const blueprint = Blueprint.argv(['echo', '{{text}}']);
      const toolFunction = Tool.createToolFunction(blueprint);

      const result = await toolFunction({});

      expect(result.content).toHaveLength(1);
      expect(result.content[0].type).toBe('text');
      expect(result.content[0].text).toBe('');
      expect(result.isError).toBe(false);
    });
  });
});
