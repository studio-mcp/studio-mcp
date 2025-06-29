import { Studio } from '../src/studio';

describe('Studio', () => {
  describe('constructor', () => {
    it('throws error with empty argv', () => {
      expect(() => new Studio([])).toThrow('usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
    });

    it('creates studio instance with valid argv', () => {
      const studio = new Studio(['echo', 'hello']);
      expect(studio).toBeInstanceOf(Studio);
    });

    it('creates studio instance with blueprint arguments', () => {
      const studio = new Studio(['echo', '{{message#Text to echo}}']);
      expect(studio).toBeInstanceOf(Studio);
    });

    it('parses --debug flag correctly', () => {
      const studio = new Studio(['--debug', 'echo', 'hello']);
      expect(studio).toBeInstanceOf(Studio);
      expect(studio.isDebugMode).toBe(true);
    });

    it('parses multiple flags before command', () => {
      const studio = new Studio(['--debug', '--other-flag', 'echo', 'hello']);
      expect(studio).toBeInstanceOf(Studio);
      expect(studio.isDebugMode).toBe(true);
    });

    it('works without debug flag', () => {
      const studio = new Studio(['echo', 'hello']);
      expect(studio).toBeInstanceOf(Studio);
      expect(studio.isDebugMode).toBe(false);
    });

    it('throws error when only flags provided', () => {
      expect(() => new Studio(['--debug'])).toThrow('usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
    });

    describe('help functionality', () => {
      let consoleInfoSpy: jest.SpyInstance;
      let processExitSpy: jest.SpyInstance;

      beforeEach(() => {
        consoleInfoSpy = jest.spyOn(console, 'info').mockImplementation();
        processExitSpy = jest.spyOn(process, 'exit').mockImplementation();
      });

      afterEach(() => {
        consoleInfoSpy.mockRestore();
        processExitSpy.mockRestore();
      });

      it('shows help and exits with --help flag', () => {
        new Studio(['--help']);

        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('usage: studio-mcp [--debug] <command>'));
        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('studio-mcp is a tool for running a single command MCP server'));
        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('-h, --help - Show this help message and exit'));
        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('--debug - Print debug logs to stderr'));
        expect(processExitSpy).toHaveBeenCalledWith(0);
      });

      it('shows help and exits with -h flag', () => {
        new Studio(['-h']);

        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('usage: studio-mcp [--debug] <command>'));
        expect(processExitSpy).toHaveBeenCalledWith(0);
      });

      it('shows help with --help flag even when command is provided', () => {
        new Studio(['--help', 'echo', 'hello']);

        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('usage: studio-mcp [--debug] <command>'));
        expect(processExitSpy).toHaveBeenCalledWith(0);
      });

      it('shows help with -h flag even when command is provided', () => {
        new Studio(['-h', 'echo', 'hello']);

        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('usage: studio-mcp [--debug] <command>'));
        expect(processExitSpy).toHaveBeenCalledWith(0);
      });

      it('shows help with --debug --help flags', () => {
        new Studio(['--debug', '--help']);

        expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('usage: studio-mcp [--debug] <command>'));
        expect(processExitSpy).toHaveBeenCalledWith(0);
      });

      it('shows complete help text content', () => {
        new Studio(['--help']);

        const helpCall = consoleInfoSpy.mock.calls[0][0];
        expect(helpCall).toContain('usage: studio-mcp [--debug] <command> --example "{{req # required arg}}" "[args... # array of args]"');
        expect(helpCall).toContain('studio-mcp is a tool for running a single command MCP server.');
        expect(helpCall).toContain('-h, --help - Show this help message and exit.');
        expect(helpCall).toContain('--debug - Print debug logs to stderr to diagnose MCP server issues.');
        expect(helpCall).toContain('the command starts at the first non-flag argument:');
        expect(helpCall).toContain('<command> - the shell command to run.');
        expect(helpCall).toContain('arguments can be templated as their own shellword or as part of a shellword:');
        expect(helpCall).toContain('"{{req # required arg}}" - tell the LLM about a required arg named \'req\'.');
        expect(helpCall).toContain('"[args... # array of args]" - tell the LLM about an optional array of args named \'args\'.');
        expect(helpCall).toContain('"[opt # optional string]" - a optional string arg named \'opt\' (not in example).');
        expect(helpCall).toContain('"https://en.wikipedia.org/wiki/{{wiki_page_name}}" - an example partially templated words.');
        expect(helpCall).toContain('Example:');
        expect(helpCall).toContain('studio-mcp say -v siri "{{speech # a concise phrase to say outloud to the user}}"');
      });
    });
  });

  describe('serve method', () => {
    it('creates and starts server without throwing', async () => {
      const studio = new Studio(['echo', 'hello']);

      // Mock the server.connect method to avoid actually starting the server
      const mockConnect = jest.fn().mockResolvedValue(undefined);
      (studio as any).server.connect = mockConnect;

      await studio.serve();

      expect(mockConnect).toHaveBeenCalledWith(expect.any(Object));
    });
  });

  describe('static serve method', () => {
    // Mock process.exit and console.error for testing
    let processExitSpy: jest.SpyInstance;
    let consoleErrorSpy: jest.SpyInstance;

    beforeEach(() => {
      processExitSpy = jest.spyOn(process, 'exit').mockImplementation();
      consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation();
    });

    afterEach(() => {
      processExitSpy.mockRestore();
      consoleErrorSpy.mockRestore();
    });

    it('handles empty argv error', async () => {
      await Studio.serve([]);

      expect(consoleErrorSpy).toHaveBeenCalledWith('Studio error: usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it('handles serve errors', async () => {
      // Mock the serve method to throw an error
      const mockServe = jest.fn().mockRejectedValue(new Error('Server connection failed'));
      Studio.prototype.serve = mockServe;

      await Studio.serve(['echo', 'hello']);

      expect(consoleErrorSpy).toHaveBeenCalledWith('Studio error: Server connection failed');
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it('handles non-Error objects', async () => {
      // Mock the serve method to throw a non-Error object
      const mockServe = jest.fn().mockRejectedValue('String error');
      Studio.prototype.serve = mockServe;

      await Studio.serve(['echo', 'hello']);

      expect(consoleErrorSpy).toHaveBeenCalledWith('Studio error: String error');
      expect(processExitSpy).toHaveBeenCalledWith(1);
    });

    it('creates and serves successfully with valid args', async () => {
      // Mock the serve method to avoid actually starting the server
      const mockServe = jest.fn().mockResolvedValue(undefined);
      Studio.prototype.serve = mockServe;

      await Studio.serve(['echo', 'hello']);

      expect(mockServe).toHaveBeenCalled();
      expect(consoleErrorSpy).not.toHaveBeenCalled();
      expect(processExitSpy).not.toHaveBeenCalled();
    });

    it('handles help flag in static serve method', async () => {
      // Mock console.info and process.exit for help functionality
      const consoleInfoSpy = jest.spyOn(console, 'info').mockImplementation();
      const processExitSpy = jest.spyOn(process, 'exit').mockImplementation();

      await Studio.serve(['--help']);

      expect(consoleInfoSpy).toHaveBeenCalledWith(expect.stringContaining('usage: studio-mcp [--debug] <command>'));
      expect(processExitSpy).toHaveBeenCalledWith(0);

      consoleInfoSpy.mockRestore();
      processExitSpy.mockRestore();
    });
  });
});
