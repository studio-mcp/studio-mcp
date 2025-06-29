import { Studio } from '../src/studio';

describe('Studio', () => {
  describe('constructor', () => {
    it('throws error with empty argv', () => {
      expect(() => new Studio([])).toThrow('Usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
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
      expect(() => new Studio(['--debug'])).toThrow('Usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
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

      expect(consoleErrorSpy).toHaveBeenCalledWith('Studio error: Usage: studio-mcp <command> --example "{{req # required arg}}" "[args... # array of args]"');
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
  });
});
