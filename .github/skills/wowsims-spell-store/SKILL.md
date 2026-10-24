---
name: wowsims-spell-store
description: 'Use when Go code reads the WoWSims Forever spell store, sim/core/spelldata: picking a rank off a ladder (Rank, Highest, ByID, Len, Contains), Find or MustFind on an id, choosing between EffectN, EffectAt and Effect(aura, misc), reading Average, FractionAt or TenthsAt, following Trigger(), Triggered() or Refs() to the spell that does the work, asking whether a spell is one ability, telling whether a class file reads rows by hand or through the resolvers, or running go run ./tools/spelldata (an id, -family, -expr, -config) or the editor hover. Use it too when a value is one rank or one effect off, Highest() names the wrong spell, a misc value names the wrong stat, a ladder reader panics on a two-effect row, a ladder misses or doubles a rank, or a Wowhead number disagrees with the row. Building a spell, aura or dot from a row is wowsims-abilities-auras; a proc or trigger is wowsims-procs; regenerating the store or the database is wowsims-db-pipeline; the map of every spell skill is wowsims-spells.'
---

# wowsims-spell-store

The store holds every spell row the sim can reach, each field the game files' column in the game files' units, and
the accessors convert on the way out. The rule this skill exists for: name the rank and the effect you mean. A rank
comes from `Rank(n)`, `ByID(id)` or an edge off another row, an effect from its position or its aura, never from
whichever one happens to give the number you expected.

Text written here follows `wowsims-conventions`. The detail lives in the pages below; this skill routes to them and holds
the traps they do not carry.

## Where things live

| Path                                          | What it is                                                                                      |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| `sim/core/spelldata/spells_auto_gen.go`       | the generated store; `sim/core` must not import `spelldata`, since the store imports core       |
| `sim/core/spelldata/store.go`                 | `Find`, `MustFind`, `All`, `ByName`                                                             |
| `sim/core/spelldata/spell.go`                 | a row: `EffectN`, `Effect`, `FindEffect`, the role readers, `Triggered`, `Refs`, `ReagentCount` |
| `sim/core/spelldata/effect.go`                | an effect's value readers: `Average`, `Percent`, `Tenths`, `Roll`, `Trigger`                    |
| `sim/core/spelldata/ladder.go`                | `Ladder`, `Ranked`, `Talent` and the per-rank readers                                           |
| `sim/core/spelldata/attributes.go`            | the attribute bits the sim acts on                                                              |
| `sim/<class>/spell_data_auto_gen.go`          | a class's `spellData`: one ladder per family, no numbers of its own                             |
| `tools/database/gen_spell_data.go`            | `discoverLadders`, which decides which rows become a family's ranks                             |
| `tools/database/overrides/spell_overrides.go` | numbers the game files do not state                                                             |
| `tools/spelldata/`                            | the command-line reader and the language server behind the editor hover                         |

| You are…                                                                               | Open                                 |
| -------------------------------------------------------------------------------------- | ------------------------------------ |
| looking up an accessor, a unit, `Nil` chaining, class flags or a ladder reader         | `docs/spell_data/store.md`           |
| reading values into a hand-built config, a borrowed tick, or the inputs a dot captures | `docs/spell_data/reading-by-hand.md` |
| running the tool, reading its JSON, or setting up the hover in an editor               | `tools/spelldata/README.md`          |
| chasing a value that is off: rage in tenths, ranks the game never grants, the rounding | `docs/spell_data/pitfalls.md`        |
| building a ladder's ranks, reading a talent's curve, or checking a row against Wowhead | `references/row-traps.md`            |

## Reading a row

```go
var slamRank = spellData.Slam.Highest()          // a ladder: Rank(n), Highest(), ByID(id), Len(), Contains(spell)
var slamDamage = slamRank.DamageEffect().Average()
rank := spellData.Defiance.Rank(points)          // rank 0 is untaken and answers Nil, so no `if rank > 0`
spelldata.Find(id)                               // Nil where the store lacks it; chains safely
spelldata.MustFind(id)                           // panics instead, for a spell the caller depends on
```

- A ladder's `ValueAt`, `FractionAt`, `MultiplierAt` and `TenthsAt` panic on any rank that does not carry exactly one
  effect. On such a row, name the effect with `EffectAt(n)` or `Effect(aura, misc)`, or let the parse read it.
- SQL array columns in `wowsims.db` are 0-based and `EffectN` is 1-based, so `EffectMiscValue_0` is `EffectN(1)`'s
  misc value.
- core's `SpellSchool` bits are the game files' bits, so a school mask is carried, not translated. `stats.SchoolIndex`
  (array positions) and `proto.SpellSchool` are separate enums.
- Ask whether a spell is one ability with `spellData.X.Contains(spell)`, on a `ProcTrigger` in its `ExtraCondition`,
  never with `ClassFlags`, which stay for talent and set filters (PR #93).

`docs/spell_data/store.md` has the rest: "Reaching an effect" for `EffectN`, `Effect(aura, misc)` and the role
readers, "Reading a value" for `Average` and the roll, "Attributes, edges and class flags" for `Trigger()`,
`Triggered()`, `Refs()` and why class flags hear several abilities. The per-level gain a `Ranked` ladder's readers skip
is in `docs/spell_data/reading-by-hand.md`.

## Hand reads and resolvers

The resolvers (`SpellConfig`, `AuraConfig`, `DotConfig`, `ParseStatic`, `ParseEffects`, `ProcTrigger`) fill what the
row states and nothing else. A hand read builds `core.SpellConfig` field by field from the row. Which way a file goes
is read off the file, so match the file you touch: `grep -l 'spelldata.SpellConfig\|spelldata.Parse' sim/<class>/*.go`.
Every class folder has files on the resolvers, and a resolver-built spell still reads some values by hand
(Consecration's ticks). The priest builds through the resolvers, not hand reads (PR #246).

A hand read (PR #42):

- reads every amount as `Average()`;
- reads `Effects[i]` as `EffectN(i+1)`, since `EffectAt(n)` is a 1-based position;
- uses a ladder's `...At` readers only on a `Talent` ladder or where the effect's per-level gain is 0;
- reads a borrowed tick through the family's `XTriggered` ladder or `Refs()[0]`;
- captures an input the cast fixes (combo points, a stack) at the cast, for `OnTick` to read from a variable.

## The spelldata tool

`tools/spelldata/README.md` has every mode, the JSON shape and the editor hover. Two modes answer the questions this
skill raises: `-expr` reads a ladder chain the way a class file writes it and marks the effect it read, and `-config`
states each field a resolver call fills and the step that filled it.

- The tool has no `-h`; it answers `unknown argument "-h"`. Run it with no arguments for the usage line.
- `go run ./tools/spelldata` needs the generated protobuf Go code, so a fresh checkout runs
  `make sim/core/proto/api.pb.go` first. The editor launcher does this itself.

## Traps

- **`Highest()` is the ladder's last entry, not its top rank.** `spellData.MutilateTriggered` runs in id order and
  ends on a rank 1, and a database update that appends rows (FlametongueWeaponTriggered gaining 29469 and 29470)
  moves it silently. Read a triggered row off the cast's top rank: `mutilateCastRank.EffectN(2).Trigger()`, or
  `Refs()`.
- **A hand read of `Misc` gets the wrong stat.** An `A_MOD_TOTAL_STAT_PERCENTAGE` effect names its stats in the second
  misc value where it sets one, a bitmask, and the first then does not count (Heart of the Wild 17003 states 0 and 8).
  `Effect.TotalStatPercentStats()` reads it, as the parse and the tool do.

The traps in the rows themselves (ladders keyed by name, granted copies, talent curves, summons, Wowhead's numbers
against the row) are in `references/row-traps.md`.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. Re-derive a claim before
repeating it:

- which files are on the resolvers: `grep -l 'spelldata.SpellConfig\|spelldata.Parse' sim/*/*.go`
- the tool's arguments: `go run ./tools/spelldata` with no arguments prints the usage line (`usage` in
  `tools/spelldata/main.go`)
- the ladder readers: `grep -n '^func (l Ladder)' sim/core/spelldata/ladder.go`
- the effect readers: `grep -n '^func (e \*Effect)' sim/core/spelldata/effect.go`

| You changed…                                                    | Update                                                                                                                    |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `sim/core/spelldata/ladder.go`                                  | "Reading a row", the `Highest()` trap, the ladder table (its `Highest()` row too) in `docs/spell_data/store.md`           |
| `sim/core/spelldata/effect.go`, `spell.go` or `store.go`        | "Reading a row", `docs/spell_data/store.md`, the `-expr` example in `tools/spelldata/README.md`                           |
| a class file moved between hand reads and the resolvers         | "Hand reads and resolvers", the opening sentences of `docs/spell_data/reading-by-hand.md` and `docs/spell_data/README.md` |
| the tool's arguments in `tools/spelldata/main.go`, or its hover | "The spelldata tool" and the command block in `tools/spelldata/README.md`                                                 |
| `discoverLadders` in `tools/database/gen_spell_data.go`         | `references/row-traps.md`                                                                                                 |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-spell-store
```

A ruling, trap or workflow fact learned while using this skill is written here in the same change that uses it. A
memory file is a pointer to this skill, not a second copy.
