# mise

A CLI that provides a suite of AI-assisted tools to help you manage your
[Tandoor](https://tandoor.dev) recipes.

## Features

- Cleans up messy ingredient rows so amount, unit, and food land in the right
  fields.

## Install

```sh
go install github.com/derethil/mise@latest
```

Or build from source:

```sh
git clone https://github.com/derethil/mise.git
cd mise
go build -o mise .
```

A Nix flake is also provided (`nix build .#mise`).

## Configuration

`mise` reads a config file from `$XDG_CONFIG_HOME/mise/config.toml` (usually
`~/.config/mise/config.toml`):

```toml
[tandoor]
token    = "your-tandoor-api-token"
base_url = "https://your-tandoor-instance/api"

[providers.ollama]
base_url = "http://localhost:11434"

[models]
small = "ollama/qwen3:8b"
large = "ollama/qwen3:14b"

[backup]
keep = 5 # number of backups to keep per recipe (0 = keep all)
```

Ollama is the only supported provider right now. `small`/`large` let you pick
different models for lighter vs. more demanding tasks. You can also configure
mise using env vars e.g. `MISE_TANDOOR_TOKEN`.

## Commands

- `mise recipe backup` / `restore` - snapshot a recipe's JSON before you mess
  with it, or roll back to a snapshot later.
- `mise recipe clean` - fixes ingredient rows that got imported as ingredients
  but aren't, and splits amount/unit/food when they've landed in the wrong
  field.
- `mise models` - list, pull, or clear the Ollama models mise uses.
- `mise logs` - print the path to the mise log file, useful when a model or
  command misbehaves.

Global flags:

- `-m, --model` - override the AI model to use for a command.
- `-v, --verbose` - enable debug logging (repeat as `-vv` for verbose output
  from the underlying AI library too).

## License

MIT, see [LICENSE](./LICENSE).
