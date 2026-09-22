# tools/spelldata

Reads the spell store, `sim/core/spelldata`, from the command line and from an editor. Run it from the
repository root.

```
go run ./tools/spelldata 11574                  # one row: header, ladder call, each effect worded and literal
go run ./tools/spelldata Whirlwind [-all]       # every row with that name, then the highest rank (or all)
go run ./tools/spelldata 11574 -json            # the same as JSON
go run ./tools/spelldata -family warrior/Execute
go run ./tools/spelldata -expr 'spellData.Execute.Highest().EffectN(1).Average(core.CharacterLevel)' -package warrior
go run ./tools/spelldata -config 'spelldata.SpellConfig(&warrior.Unit, executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))' -package warrior
go run ./tools/spelldata -hover sim/warrior/execute.go 12:40   # the markdown an editor shows at line:column (1-based)
go run ./tools/spelldata -lsp                   # a language server on stdio
```

`-hover` prints the hover's markdown on stdout and its trace on stderr, and exits 1 where there is no
hover. `-config` resolves a `spelldata.SpellConfig` call - the row pick substituted through the
package's declarations, each option applied - and states each field with the step that filled it.

## The language server

`-lsp` speaks JSON-RPC 2.0 over stdio with Content-Length framing.

| method                                                |                                                                                                             |
| ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `initialize`                                          | `hoverProvider`, full-text `textDocumentSync`; `initializationOptions.trace` is `"on"` (default) or `"off"` |
| `initialized`, `$/cancelRequest`, other notifications | ignored                                                                                                     |
| `textDocument/didOpen`, `didChange`                   | keeps the buffer, so a hover reads unsaved edits; drops the folder's declaration cache                      |
| `textDocument/didSave`, `didClose`                    | drops the folder's declaration cache (and the buffer, on close)                                             |
| `textDocument/hover`                                  | markdown `MarkupContent`, or `null`                                                                         |
| `shutdown`, `exit`                                    | exits 0 after a `shutdown`, 1 without one                                                                   |
| any other request                                     | `MethodNotFound`                                                                                            |

Every hover sends its trace as `window/logMessage` of type Log: the position, what matched (`id`,
`family`, `ident`, `segment`, `SpellConfig`), each substitution with the file and line it came from,
the declaration cache hit or miss, and the result (`✓ 20662 Execute (Rank 5) → effect 1 → 600`) or why
it stopped (`✗ ...`).

What a hover reads: a spell id in the shapes hand-written code states (`MustFind(n)`, `Find(n)`,
`.ByID(n)`, `SpellID: n`, `"spellId": n`, `fromSpellId(n)`, `spellId: n`) in any file; in a Go file
also a `spellData.<Family>` token, a name the package folder binds to a ladder chain (substituted
through at most four names, cycles refused), one accessor of a chain on its declaration line, and
`spelldata.SpellConfig`. The server compiles the store in; restart it after regenerating.

## Editors

- **VS Code**: `.vscode/extensions/wowsims-spelldata`, a workspace extension built from
  `tools/vscode-spelldata` by `make vscode-spelldata` (or `npm run build:lsp` from the root, which also compiles the server).
- **Zed**: `tools/zed-spelldata`, installed with `zed: install dev extension`.

### Other editors

Neovim (0.10 or later), in `init.lua`:

```lua
vim.api.nvim_create_autocmd('FileType', {
	pattern = { 'go', 'json', 'typescript', 'typescriptreact' },
	callback = function(args)
		local root = vim.fs.root(args.buf, 'go.mod')
		if root and vim.uv.fs_stat(root .. '/tools/spelldata') then
			vim.lsp.start({
				name = 'wowsims-spelldata',
				cmd = { 'go', 'run', './tools/spelldata', '-lsp' },
				cmd_cwd = root,
				root_dir = root,
				init_options = { trace = 'on' },
			})
		end
	end,
})
```

Helix, in `languages.toml` (the repository's `.helix/languages.toml` or the user's):

```toml
[language-server.wowsims-spelldata]
command = "go"
args = ["run", "./tools/spelldata", "-lsp"]
config = { trace = "on" }

[[language]]
name = "go"
language-servers = ["gopls", "wowsims-spelldata"]

[[language]]
name = "json"
language-servers = ["vscode-json-language-server", "wowsims-spelldata"]

[[language]]
name = "typescript"
language-servers = ["typescript-language-server", "wowsims-spelldata"]

[[language]]
name = "tsx"
language-servers = ["typescript-language-server", "wowsims-spelldata"]
```

Helix starts the server in the root it finds for the file, which in this repository is its root - where
`go run ./tools/spelldata` has to run. Its hover shows every server's answer.
