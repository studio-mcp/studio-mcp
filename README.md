# Studio MCP

Make any CLI into a single tool MCP server.

> Bright Studio Apt – No Walls, No Rules – $0/mo OBO.
>
> Wired and wild. No questions. rm -rf compatible. Cash only. Basement-ish. Just enough, nothing more.
>
> Text 404 to /dev/null for more details!

## What's Included?

`studio-mcp` is the simplest possible StdIO transport Model Context Protocol server.

Everything after the `studio-mcp` command will be turned into an MCP tool that runs just that command when called by Cursor, Claude, etc.

`studio-mcp` uses a very simple Mustache-like template syntax in order to tell the LLM how to use your MCP command.

```sh
$ npx -y studio-mcp command "{{ required_argument # Description of argument }}" "[optional_args... # any array of arguments]"
```

`studio-mcp` turns this into an input schema for the MCP tool so that tool calls know what to send:

```json
{
  "type": "object",
  "properties": {
    "required_argument": { "type": "string", "description": "Description of argument" },
    "optional_args": { "type": "array", "items": { "type": "string" }, "description": "any array of arguments" }
  },
  "required": ["required_argument"]
}
```

You can run almost any command. Since you're just renting the place, please be a good tenant and don't `rm -rf` anything.

## Install

If you don't have a modern node installed as your system node:

```sh
brew install node
```

After that, you can install this as an mcp server:

### Claude Desktop

We'll use `echo` as an example command. It's almost perfectly useless, but it's easy to understand.

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "echo": {
      "command": "studio-mcp",
      "args": ["echo", "{{text#What do you want to say?}}"]
    },
    "git-log": {
      "command": "studio-mcp",
      "args": ["git", "log", "--oneline", "-n", "20", "{{branch#Git branch name}}"]
    }
  }
}
```

### Cursor

Add to your `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "echo": {
      "command": "npx",
      "args": ["-y", "studio-mcp", "echo", "{{text#What do you want to say?}}"]
    },
  }
}
```

(this is a very useless example, but it will make sure it's working)

### VSCode

```json
{
  "mcp": {
    "servers": {
      "echo": {
        "type": "stdio",
        "command": "studio-mcp",
        "args": ["echo", "{{text#What do you want to say?}}"]
      }
    }
  }
}
```

## Blueprint Syntax

Studio uses blueprints (templates) to keep your studio tidy.

```bash
studio-mcp rails generate "{{generator#A valid rails generator like scaffold, model}}" "[args...#Any additional args needed for the generator]"
```

This creates a Studio server with two arguments: `generator` and `args`.
Everything after the `#` will be used as the description for the LLM to understand.

Blueprints use the format: `{{name # description}}` and `[name # description]` for string and array arguments.

- `{{name}}`: Required string argument
- `[name]`: Optional string argument
- `[name...]`: Required array argument (spreads as multiple command line args)
- `name`: The argument name that will be exposed in the MCP tool schema. May not contain spaces.
- `description`: A description of what the argument should contain. May contain spaces.

This is a simple studio, not one of those fancy 1 bedroom flats.
Blueprint types, flags, validation??? The landlord will probably upgrade the place for free eventually... right?

## Development

To build and test locally:

```bash
npm install
npm run build
npm run dev -- echo "{{text#What to echo?}}"
```

### Testing

The project includes comprehensive tests covering:
- Blueprint parsing and schema generation
- Command execution and error handling
- MCP protocol integration
- Real-world usage scenarios

```bash
npm test              # Run all tests
npm run test:watch    # Run tests in watch mode
npm run test:coverage # Run tests with coverage report
```

Test coverage is excellent with 87% statement coverage including:
- 100% Blueprint functionality
- 95% Tool execution logic
- Full integration testing of MCP protocol

## Home Is Where You Make It

This is your studio too. Bugs, features, ideas? Swing by the repo:

🏠 https://github.com/martinemde/studio

## Lease Terms: MIT

Move in under standard terms, no fine print. Full text here: [MIT License](https://opensource.org/licenses/MIT).
