---
name: wowsims-db-pipeline
description: 'Use when WoWSims Forever work runs or repairs the generators behind committed game files data: make db, make go-to-ts, gen_spelldata, gen_db, db2tool; regenerating the store or the database; adding or removing an override in tools/database/overrides or overrides.go, the numbered .sql among them; a local wowsims.db behind master; merging master into a branch that conflicts in spell_store_inputs.json, spells_auto_gen.go, db.json or db.bin; the Check Hotfixes and Update Items workflows; the proc audit; a table wowsims.db does not carry. Use it too when a generator stops on hotfixes or a stale override, -check names a file, a regeneration deletes items master has, db.bin stays behind db.json, unsupported_procs.txt moves, or a value seems missing from the sim. Reading a row is wowsims-spell-store; item and consumable effects are wowsims-consumables-items; proc wiring is wowsims-procs; proto field rules are wowsims-core-mechanics; tests and .results are wowsims-testing; the spell skills map is wowsims-spells.'
---

# wowsims-db-pipeline

Every committed file derived from the game files has a generator, and the generator writes it: the spell store and
class ladders (`gen_spelldata`), the item, enchant and consumable database with its registrations (`gen_db`), and the
TS copies of Go tables (`gen_db -gen=go-to-ts`). The rule this skill exists for: never write generated output by hand,
and never regenerate from a database that is behind what master committed. Where the local database cannot match
master, splice your own change into master's files and prove the splice. Text here follows `wowsims-conventions`; the
pages below hold the detail, and this skill carries the traps and commands they lack.

## Where things live

| Path                                                                        | What it is                                                                                       |
| --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------ |
| `tools/database/wowsims.db`                                                 | the game files database, gitignored, written by `tools/db2tool/` from `generator-settings.json`  |
| `tools/database/generator-settings.json`                                    | the tables `db2tool` extracts; `ptr-generator-settings.json` differs in `Product` only           |
| `tools/database/gen_spelldata/`                                             | the store, form and class ladder generator, a separate binary from `gen_db`                      |
| `tools/database/gen_db/main.go`                                             | `gen_db`: the hotfix check, the numbered `.sql`, the item database, registrations, talent files  |
| `tools/database/overrides/`                                                 | spell, weapon and enchant rate overrides, hand triggers, extra spells, and the numbered `.sql`   |
| `tools/database/overrides.go`                                               | item, enchant and consumable overrides, allow and deny lists, `ItemSetClasses`                   |
| `assets/db_inputs/spell_store_inputs.json`                                  | the rows the store was built from, with each group's `Builds` and `Hotfixes`; gzipped (`zcat`)   |
| `assets/database/db.json`, `assets/database/db.bin`                         | one `proto.UIDatabase` written twice by `gen_db`; the sim embeds the binary, the UI fetches JSON |
| `.github/workflows/check_hotfixes.yml`, `update_items.yml`, `update_db.yml` | the hourly hotfix check, the items-only regeneration it dispatches, the full one run by hand     |

| You are…                                                                       | Open                                                               |
| ------------------------------------------------------------------------------ | ------------------------------------------------------------------ |
| running a generator, `-check` or `-unchecked`, or reading what guards them     | `docs/spell_data/regenerating.md`                                  |
| taking a class, the racials or the items onto a build, or stamping `BuildInfo` | `docs/spell_data/partial-updates.md`, its §Commands for every flag |
| adding a number the game files do not state                                    | `docs/spell_data/overrides.md`, then "Overrides" below             |
| adding, removing or correcting a consumable                                    | `docs/consumables.md`, then `wowsims-consumables-items`            |
| extracting a table, or a `db2tool` flag                                        | `tools/db2tool/README.md`                                          |
| running the class suites or checking `.results`                                | `wowsims-testing` `## The gates`                                   |

## Regenerating

A full regeneration runs in this order, each step reading what the one before wrote (`update_db.yml`):

1. `make db`: `db2tool` extracts `wowsims.db`, then `gen_spelldata`, then `gen_db -gen=db`;
2. `make basestats`, which reads `wowsims.db` into `sim/core/base_stats_auto_gen.go`;
3. `gen_spelldata` once more, then `make proto`, since `gen_db` rewrites the talent enums in `proto/`;
4. `go run ./tools/database/gen_db -gen=go-to-ts` last, since the TS constants come from the regenerated Go.

```
go run ./tools/database/gen_spelldata                        # the store, the forms, every class file
go run tools/database/gen_db/*.go -outDir=./assets -gen=db   # db.json, db.bin, registrations, talent files
go run ./tools/database/gen_spelldata -classes=none -shared=false && \
  go run tools/database/gen_db/*.go -outDir=./assets -gen=db -classes=none   # the items alone, no re-extraction
go test ./sim/core/spelldata/ ./tools/database/ -count=1     # no database needed
```

- **A checkout keeps its own `wowsims.db`.** `gen_db` runs every numbered `.sql` in `tools/database/overrides/`
  against the database it opens, in numeric order, before it reads a row, so a database shared with another checkout
  changes under it.
- **Never delete a generated Go file before regenerating it.** `gen_db` imports the sim, so a class package missing
  its file breaks the generator that would recreate it.
- **`make db` re-extracts.** For the items alone from the database you have, run the two generators above; a new
  extraction can take hotfixes the committed rows do not hold.
- **Edit the Go table, never its TS copy.** `make go-to-ts` writes the files in `AUTO_GEN_FILES_TS` from Go
  constants; the EP pseudo-stat list and `getRootStat` stay hand-kept in TS. On a fresh checkout `make go-to-ts`
  takes its outputs as up to date, so call `gen_db -gen=go-to-ts` directly.
- **A new parse row moves generated output.** It flips `Effect.ParsedAura()` for every misc value, which
  `wowsims-consumables-items` "The pipeline" describes. Regenerate at the base and on the branch, diff the two, and
  grep the hand-written code for the same effect before shipping the row.
- **The proc audit list is a gitignored baseline its own test writes.** After new spell rows arrive, rewrite
  `tools/database/unsupported_procs.txt` with `UPDATE_UNSUPPORTED_PROCS=1` on `TestEveryReachableProcIsSupportedOrListed`
  instead of reading its diff as a regression; `UPDATE_UNSUPPORTED_CONSUMABLES` does the same for the consumables.

## Overrides

`docs/spell_data/overrides.md` holds each field, the rule that makes it stale, and the hand triggers and extra
spells. The generator refuses an override once the game files state the value; remove the row then, never the check.

- **Read the overrides before claiming the sim lacks a value.** `wowsims.db` holds only the game files' values; the
  store also carries the hand-kept ones, each marked `// override:` in `spells_auto_gen.go`.
- **The table follows whose number it is.** A rate that belongs to one weapon goes in `WeaponProcs`, keyed by item and
  spell, since two weapons can share a trigger spell; an enchant's in `EnchantProcs`; a rate every carrier of the
  spell shares (Omen of Clarity, Seal of Light) in `Spells` with `Field: PPM`. Proc wiring is `wowsims-procs`.
- **Rows are keyed struct literals, one field per line.** A value nobody measured on Forever carries
  `// TODO: <origin> value - verify in-game` on its value line, and `Source` names the origin.

## A database behind master

The generators refuse a database older in hotfixes than the committed inputs (`RequireCommittedHotfixes`,
`docs/spell_data/partial-updates.md`), since a regeneration from it deletes every row only a later hotfix adds. Match
master's database first, from the cache the Update Items pull request names; when it cannot match, splice your own
change into master's files and prove the splice. Both, and merge conflicts in the store, `db.json` and `db.bin`, are
in `references/behind-master.md`. Proto field numbers on a merge are `wowsims-core-mechanics` `## Protos`.

## Hotfixes

The workflows' header comments and the end of `docs/spell_data/partial-updates.md` §Commands describe them, except:

- **A cache is dispatched once.** Check Hotfixes reads raidbots' verified list and fires only when the newest cache is
  for the build the CDN serves and its URL is not yet in the actions cache.
- **The tests run inside Update Items** (`run_tests.yml`), since a pull request the workflow token opens gets no checks.
- **The merge job dispatches Build and Deploy only while its state reads `active`.** Read it before a dispatch by
  hand: `gh api repos/wowsims/forever/actions/workflows/deploy.yml --jq .state`.

## Tables the generator does not ship

`wowsims.db` holds only the tables in `generator-settings.json`'s `Tables`; a name added there is extracted on every
run. To read another table once, copy the settings file with only that table listed and run
`db2tool -s <copy> --output <other.db>`; never re-extract into the working `wowsims.db`. To add one table to the
working database, `ATTACH` the other and `CREATE TABLE … AS SELECT`. A table missing from the build aborts the run,
and one the install lacks (`not found in local CASC indices`) comes from `db2tool --cdn --dbcache <DBCache.bin>`.
What the tables hold and how a script reads them: `references/game-files-tables.md`.

## Traps

- `MIN_EFFECT_ILVL = 40` in `tools/database/gen_effects.go` drops chance-on-hit items below item level 40 (Bloodpike).
- `InferPhase` in `tools/database/item_source_utils.go` has no callers and stays: `gen_db` stamps every item phase 1.
- A test that assumes one build across every spell data group breaks once a partial update mixes the builds.
- `make buffs-proto` needs the compiled protos; on a fresh checkout run `make sim/core/proto/api.pb.go` first.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. Re-derive a claim before
repeating it:

- the make targets: `grep -n '^[a-z-]*:' makefile | grep -E 'db|spelldata'`
- the workflow steps: `grep -n 'name:' .github/workflows/check_hotfixes.yml .github/workflows/update_items.yml`
- the full order: `grep -n 'make \|go run' .github/workflows/update_db.yml`
- `gen_db`'s order: `grep -n 'RequireCommittedHotfixes\|RunOverrides\|WriteBinaryAndJson' tools/database/gen_db/main.go`
- the deploy gate: `grep -n 'deploy.yml' .github/workflows/update_items.yml`
- the go-to-ts outputs: `grep -n '^AUTO_GEN_FILES_TS' makefile`
- the override tables: `grep -n '^var ' tools/database/overrides/*.go`
- the item level floor and the phase: `grep -n 'MIN_EFFECT_ILVL = \|Phase = 1' tools/database/gen_effects.go tools/database/gen_db/main.go`

| You changed…                                                               | Update                                                                                             |
| -------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| a make target, or the step order in `update_db.yml`                        | "Regenerating"                                                                                     |
| a generator flag or the group rules                                        | `docs/spell_data/partial-updates.md` §Commands, "Regenerating"                                     |
| an override field, table or stale rule                                     | `docs/spell_data/overrides.md`, "Overrides"                                                        |
| `RequireCommittedHotfixes`, `WriteBinaryAndJson` or `ReadDatabaseFromJson` | "A database behind master", `references/behind-master.md`                                          |
| `check_hotfixes.yml` or `update_items.yml`                                 | "Hotfixes", the Update Items paragraph in `docs/spell_data/partial-updates.md`                     |
| `generator-settings.json` or `db2tool`'s flags or schema                   | "Tables the generator does not ship", `references/game-files-tables.md`, `tools/db2tool/README.md` |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-db-pipeline
```

A ruling, trap or workflow fact learned while using this skill is written here in the same change that uses it. A
memory file is a pointer to this skill, not a second copy.
