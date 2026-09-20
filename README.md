# mise

A CLI that provides a suite of AI-assisted tools to help you manage your
[Tandoor](https://tandoor.dev) recipes. LLMs are not perfect, and mise will
certainly get some things wrong, but if you have hundreds or thousands of
recipes, it can help you manage them semi-autonomously.

## Features

- Normalize messy ingredient rows so amount, unit, and food land in the right
  fields, reusing Tandoor's existing foods and units.
- Assign keywords to a recipe following your own custom keyword schema.
- Manage your local AI models, tracking downloads, disk usage, and unused
  weights for Ollama.

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

`mise` reads a config file from `$XDG_CONFIG_HOME/mise/config.toml` - see
[the example config](./docs/EXAMPLE_CONFIG.toml) for an example. You can also
configure mise using env vars e.g. `MISE_TANDOOR_TOKEN` or with a global flag
e.g. `--tandoor.token`.

### Providers

Ollama is the only supported provider right now. `small`/`large` let you pick
different models for lighter vs. more demanding tasks or you can provide a
--model flag to override the model for one run.

## Commands

- `mise configure` - manage configuration file
- `mise recipe normalize` - cleans up a recipe's ingredients by ensuring each
  amount/unit/food lands in the correct Tandoor field
- `mise recipe keyword` - assigns keywords to a recipe according to the keyword
  schema in your schema file
- `mise recipe backup` / `restore` - snapshot a recipe's JSON before you mess
  with it, or roll back to a snapshot later.
- `mise models` - manage the Ollama models mise uses
- `mise logs` - print the path to the current mise log file

Global flags:

- `-m, --model` - override the AI model to use for a single command
- `-v, --verbose` - enable debug logging (`-vv` for verbose output from the
  underlying AI library)
- `-c, --config` - override config file path

## Keyword schemas

`mise recipe keyword` has no built-in keyword assignment system, so you can tell
mise to follow whatever schema you prefer. It reads your schema file
(`$XDG_CONFIG_HOME/mise/keyword_schema.md` by default), injects it into the
prompt, and uses it to decide which keywords to add to a recipe.

```markdown
## Cuisine

Assign the broad cuisine (Chinese, Japanese, American, French) and, when the
dish belongs to a distinct regional cooking tradition, that region as well
(Szechuan, Cajun, Sicilian). A region never appears without its broad cuisine.

## Dish type

Assign exactly one: Main, Side, Bread, Appetizer, Breakfast, Dessert, or Drink.
```

Each `##` heading is one "keyword category". You can tell the model to return
multiple keywords for a category if needed but mise makes a separate model call
per category to improve accuracy. Anything above the first heading is included
in every call. A schema with no headings is simply one call.

Existing Tandoor keywords are also shown to the model so it typically reuses
them instead of creating a similar keyword.

A more fully-featured keyword classification system can be seen in
[this example schema file](./docs/EXAMPLE_KEYWORD_SCHEMA.md).

## License

MIT, see [LICENSE](./LICENSE).
