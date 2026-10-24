---
name: wowsims-spells
description: 'The map of the WoWSims Forever spell data skills. Use it when work touches game files data or the generators that write it and the owning skill is not yet clear: a spell, talent, aura, dot or proc built from its row, a consumable or item effect, an override, a regeneration, a rotation, a test, or deciding whether a sim number or the game files are wrong. Use it even when the task only says "fix this potion", "this talent is off" or "regenerate the DB", then load the skill it names. Reading a row is wowsims-spell-store; building an ability, aura, dot or talent is wowsims-abilities-auras; a proc is wowsims-procs; consumables and item effects are wowsims-consumables-items; regenerating and overrides are wowsims-db-pipeline; what core does around a cast is wowsims-core-mechanics; a rotation is wowsims-apl; proving a change is wowsims-testing; judging a diff is wowsims-review; porting a whole class is wowsims-spec-implementation.'
argument-hint: 'Describe the spell, talent, proc, consumable, item effect or generator task.'
---

# WoWSims Forever game files data

Text written here follows `wowsims-conventions`. This skill is the map: find the task under "Which skill" and load
that skill. The detail lives in `docs/spell_data/` (start at its README), `docs/consumables.md` and
`tools/spelldata/README.md`, which the subject skills route to.

## Where things live

| Path                                                 | What it is                                                                                                                                      |
| ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `sim/core/spelldata/spells_auto_gen.go`              | Generated store: every spell the sim can reach, in the game files' units.                                                                       |
| `sim/core/spelldata/*.go`                            | Hand-written accessors (`store.go`, `spell.go`, `effect.go`, `ladder.go`, `attributes.go`) and resolvers (`resolve_*.go`, `parse_effects*.go`). |
| `sim/<class>/spell_data_auto_gen.go`                 | Generated: one `spelldata.Ladder` per family, no numbers of its own.                                                                            |
| `tools/database/overrides/spell_overrides.go`        | Numbers the game files do not state, each with a reason and a rule that refuses it once the game files catch up.                                |
| `tools/database/overrides.go`                        | Item, enchant and consumable overrides, allow and deny lists.                                                                                   |
| `assets/db_inputs/spell_store_inputs.json`           | The rows the store was built from. Lets the store be regenerated and tested with no game files database.                                        |
| `assets/database/db.json`, `db.bin`, `leftover_db.*` | The item, enchant and consumable database `gen_db` writes; `db.bin` is the same `proto.UIDatabase` encoded.                                     |
| `sim/common/forever/*_auto_gen.go`                   | Generated item, enchant and consumable registrations; hand-written effects sit beside them (`items_*.go`, `consumables_*.go`).                  |
| `sim/common/shared`                                  | The builders that run consumables and item effects from their rows.                                                                             |

## Which skill

| Task                                                                                            | Skill                                                             |
| ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| reading a rank, effect or value off the store, `go run ./tools/spelldata`, hand reads           | `wowsims-spell-store`                                             |
| a number looks wrong at rank N, or a value equals another rank's                                | `wowsims-spell-store` (its traps), then `wowsims-abilities-auras` |
| building an ability, aura, dot, talent, raid buff or debuff from its row                        | `wowsims-abilities-auras`                                         |
| a proc trigger, proc chance, ICD, charged buff, item or enchant proc, a measured rate           | `wowsims-procs`                                                   |
| a consumable, an item, set or enchant effect, a `db.json` entry, an item the request must carry | `wowsims-consumables-items`                                       |
| regenerating the store or the database, overrides, a database behind master, hotfixes           | `wowsims-db-pipeline`                                             |
| what core does around a cast: order, costs, scaling, targets, hooks, auras, hit tables, protos  | `wowsims-core-mechanics`                                          |
| a rotation, an APL value or action, a preset                                                    | `wowsims-apl`                                                     |
| the gates, `.results` files, where a test goes, printing a value at runtime                     | `wowsims-testing`                                                 |
| judging a diff, your own before a commit or someone else's pull request                         | `wowsims-review`                                                  |
| porting a whole class or working a `[Spec]` issue                                               | `wowsims-spec-implementation`                                     |

## When a sim number disagrees with the game files

The rows are a second opinion on every number the sim already had. Decide rather than assume, and when a
disagreement stays, say so at the call site.

- The sim is likely wrong when its value equals a different rank's or a sibling spell's, or breaks the
  die-sides ladder of the other ranks.
- The sim is likely right when the offset is uniform across unrelated abilities: that usually means the
  sim models something the game files' base excludes (a ranged-weapon component in every hunter cast
  time), not a run of slips. A coefficient read from another spell is not a disagreement either: a tick's
  coefficient belongs to the spell its description names.
- Where the row and the tooltip disagree, the tooltip can win (Flurry's haste ladder against its buff
  row's flat 30), and the comment names the wording it came from.
