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

## Move-Install

These install instructions are like my lease agreement: full of gotchas.
Have your lawyer read it over. (You do have a lawyer right?)

If you're lucky, you can just install it:

```sh
npm install -g studio-mcp
```

After that, you'll be lucky if you can get this running without a full path somewhere. Cursor and Claude don't run in your shell.

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
studio-mcp npx -y "{{node_pkg#A somewhat crazy thing to do}}" "[args...#Any additional args needed for pwning your system]"
```

This creates a Studio server with two arguments: `node_pkg` and `args`.
Everything after the `#` will be used as the description for the LLM to understand.

Blueprints use the format: `{{name # description}}` and `[name # description]` for string and array arguments.

- `{{name}}`: Required string argument
- `[name]`: Optional string argument
- `[name...]`: Opyional array argument (spreads as multiple command line args)
- `name`: The argument name that will be shown in the MCP tool schema. Only letter numbers and underscores (dashes become underscores, case-insensitive).
- `description`: A description of what the argument should contain. May contain spaces.

This is a simple studio, not one of those fancy 1 bedroom flats.
Blueprint types, flags, validation? The landlord will probably upgrade the place for free eventually... right?

## Utilities Included

To build and test locally:

```bash
npm install
npm run build
npm run dev -- echo "{{text#What to echo?}}"
```

### Did something break?

The landlord definitely takes care of the place:
- more than none tests
- files! lots of 'em!
- maybe even some test coverage
- you still need proof of renters insurance

```bash
npm test              # Run all tests
npm run test:watch    # Run tests in watch mode
npm run test:coverage # Run tests with coverage report
```

Uncovered portions are the tenant's responsibility. (no one ever understand how hard it is for us landlords)

## Home Is Where You Make It

This is your studio too. Bugs, bedbugs, features, ideas? Swing by the repo during open-house:

🏠 https://github.com/martinemde/studio-mcp

## Lease Terms: MIT

Move in under standard terms, no fine print. Full text here: [MIT License](https://opensource.org/licenses/MIT).
