# WoWSims Spelldata for Zed

The spell hover of `tools/spelldata` in Zed: a spell id in Go, JSON, TypeScript or TSX, a ladder family,
a name bound to a ladder chain or a `spelldata.SpellConfig` call hovers as the client's row, the value
the chain reads or the config the resolver builds. The extension only starts the server,
`go -C <worktree root> run ./tools/spelldata -lsp`; everything else is in `tools/spelldata` (see its
README).

## Install

Zed builds a dev extension itself, which needs Rust installed through `rustup`:

1. Run `zed: install dev extension` from the command palette.
2. Pick this folder, `tools/zed-spelldata`.

Open the repository root as the worktree - the server is run from there and reads the store it
compiles. The first hover waits for `go run` to compile, a few seconds. `go` has to be on the PATH of
the shell Zed starts in.

The server logs each hover's trace as a log message: open _debug: open language server logs_ and pick
`wowsims-spelldata` to read it. To silence it, set the server's initialization option in Zed's
settings:

```json
{
	"lsp": {
		"wowsims-spelldata": {
			"initialization_options": { "trace": "off" }
		}
	}
}
```

The extension is registered next to Zed's own servers for those languages, so gopls and the TypeScript
server keep answering too; Zed shows both hovers.
