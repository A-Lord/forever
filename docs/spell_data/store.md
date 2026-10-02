# The store

`sim/core/spelldata/spells_auto_gen.go` holds every spell the sim can reach: the ones the class files
name, the ones items, enchants and set bonuses cast, and everything those reach in turn through a
trigger effect, an actionbar override or a tooltip reference. `sim/core/spelldata/snapshot_test.go`
pins what that comes to - 7048 rows carrying 9651 effects - so a regeneration that moves the universe
says so there.

Every field is the game files' column in the game files' units: a percentage is the integer 16, rage is on a
0-1000 bar, times are milliseconds. The conversion is in the accessors, so a row always matches what
the DBC says. `sim/core` must not import the package: the store imports core, and the import back
would be a cycle. That is why the raid buffs live in `sim/core/buffs` - see
[Buffs and debuffs](buffs-and-debuffs.md).

## Finding a row

```go
frostbolt := spelldata.Find(116)      // Nil where the store does not carry the id
frostbolt = spelldata.MustFind(116)   // panics there instead
```

Every accessor answers on `Nil`, and `Nil` chains: `spelldata.Find(0).EffectN(1).Average()` is 0
rather than a crash, so a caller can ask about a spell this build does not have. `MustFind` is the
other bargain, and it is what a package-level variable takes - a regeneration that drops an id then
fails at startup, naming the id, instead of registering a spell with no numbers. `All()` and
`ByName(name)` exist for tests and debugging; a sim names its spells by id.

An id in hand-written code reads without grepping the table: `go run ./tools/spelldata 11574` prints the
row's header and one line per effect, each worded and then followed by the game files' own columns. The
header names the school, the times and the cost, and also the stances by name, the weapon or slot the
row requires, the range, the target cap, the stacks and charges, the internal cooldown, the attributes
`attributes.go` exposes, the proc's rate and flags, the spells the tooltip references and the row's
labels. The title line names the ladder a class file reaches the id through - `warrior
spellData.Rend.Highest()` - scanned out of `sim/*/spell_data_auto_gen.go`, and says nothing for a class
the generator has not put on ladders yet. The same scan reads the other way too: `-family
warrior/Execute` prints a ladder's ranks with the call that reaches each, then its highest rank in full,
and `-expr 'spellData.Execute.Rank(3)' -package warrior` prints the row one ladder call names. A pick
followed by the store's own accessors is read too - `-expr
'spellData.Execute.Highest().EffectN(1).Average()'` answers the value, the doc comment
that accessor carries and the row with the effect it read marked `(read)`.

The worded line states each number in the unit the sim spends it in, which is also how it names the
accessor it was read with: a plain amount is `Average(60)` and a range is `Min`/`Max`, a percentage is
`Percent()`, rage is `Tenths()`, a time is `TimeValue()`, and a share of spell or attack
power is `Coeff()`/`APCoeff()`. A dummy line states the row's number and nothing about what it means,
which is the one shape that always needs the tooltip. An effect whose type or aura the printer has no
wording for says `unrecognised shape`, and the literal beneath it is then the whole answer.

A name instead of an id lists every row carrying it and then the highest rank, and
`-json` is the same answer for a tool to read. `-config 'spelldata.SpellConfig(&warrior.Unit,
executeRank, spelldata.Melee(core.ProcMaskMeleeMHSpecial))' -package warrior` prints the config the
resolver builds for that call, each field with the step that filled it - the row or the option - and
the row pick substituted through the package's declarations.

The same reading is an editor hover. `go run ./tools/spelldata -lsp` is a language server on stdio: a
hover on an id Go, APL JSON and TS state - `MustFind(11574)`, `SpellID: 11574`, `"spellId": 11574`,
`fromSpellId(23563)` - shows the row, with the header and the effects as tables; in Go a
`spellData.<Family>` token shows the ladder, a name bound to a ladder chain shows the rank or the value
it reads (substituted through the names it stands on, read from the open buffers first), each accessor
on a declaration line answers for itself, and the cursor on `spelldata.SpellConfig` shows the resolved
config. Each hover logs how it was resolved as a `window/logMessage`. `-hover sim/warrior/execute.go
12:40` prints the markdown a hover at that line and column shows, for scripts and tests. The server
compiles the store in, so a hover matches the checkout it was started in.

`.vscode/extensions/wowsims-spelldata` is its VS Code client, a workspace extension VS Code offers to
install when the repository is opened. It is built, not edited: the source is
`tools/vscode-spelldata`, and `make vscode-spelldata` bundles it with `vscode-languageclient` and writes
the manifest; `npm run check` there fails where the committed output is not a fresh build.
`tools/zed-spelldata` is the Zed extension (`zed: install dev extension`), and
`tools/spelldata/README.md` has the method table and the Neovim and Helix snippets.

## Reaching an effect

```go
damage := frostbolt.EffectN(2)                                 // the second effect the row carries
slow := frostbolt.Effect(dbcenums.A_MOD_DECREASE_SPEED, 0)     // the one effect with this aura and misc value
```

`EffectN` counts from **1, by position**. That is not the game files' `EffectIndex`, which has gaps - 46
of the store's rows state an index that is not the position it sits at - and the row keeps the game files'
own number in `Effect.Index`. A position the row does not have answers `NilEffect`.

`Effect(aura, misc)` names an effect by what it does rather than by where it sits. It panics when no
effect matches, and when two do: reading the first of several silently is the bug it exists to prevent,
and `EffectN` is the way past it. `FindEffect(type, aura, misc)` is the same question without the
panic, answering `NilEffect` instead.

The role readers each answer the first effect of their kind, or `NilEffect`: `DamageEffect()` (school
damage or any of the weapon-damage effects), `HealEffect()`, `EnergizeEffect()` and `PeriodicEffect()`
(the aura application whose aura ticks).

## Reading a value

|                         |                                                                       |
| ----------------------- | --------------------------------------------------------------------- |
| `BaseValue()`           | the game files' own number, unconverted                               |
| `Percent()`             | over 100, because the game files state a percentage as the integer 16 |
| `Tenths()`              | over 10, because rage sits on a 0-1000 bar                            |
| `TimeValue()`           | the number read as milliseconds, which is what a duration modifier is |
| `Period()`              | `EffectAuraPeriod`, the tick interval                                 |
| `Coeff()` / `APCoeff()` | the spell power and attack power shares of the effect                 |
| `Average()`             | the amount at the sim's character level                               |
| `Min()` / `Max()`       | the ends of the roll                                                  |
| `Mean()`                | what a cast averages, the middle of the roll                          |
| `Roll(sim)`             | the amount for one cast                                               |

Each also comes in an `At` form taking a level (`AverageAt(level)`, `RollAt(sim, level)`), for a test
pinning an in-game tooltip at a lower level and for `ParseEffects`' level option.

`AverageAt(level)` folds the amount the way the tooltip does: the base points plus
`EffectRealPointsPerLevel` for each level between the spell's `BaseLevel` and the caster's, stopped
at `MaxLevel` where the row states one, and floored. `BaseLevel` equals `SpellLevel` on most ranks;
where it does not - Vindication's debuff states 0 and 1 - the in-game amount counts from `BaseLevel`.
The arithmetic is float32 on purpose - the game files' per-level gain is a float32 widened into the
database, and folding it in float64 moves six rows off the tooltip.

`EffectVariance` is the spread the amount rolls over. It spreads the base points only:
`|BasePoints| * Variance/2` either side of them, each end rounded to a whole number, then the per-level
gain is added to both and `Min` is rounded down and `Max` rounded up, in float32 for the same reason,
with a base end on a half rounding to the even one. `Roll` rolls it where the row states one and
answers the average where it does not, so a call site never has to ask whether this particular spell's
amount is a range.

## Attributes, edges and class flags

`Spell.Attr` is the game files' 17 attribute words. `HasAttr(word, bit)` reads one - a word past the end
answers unset rather than panicking - and `sim/core/spelldata/attributes.go` names the ones the sim
acts on: `IsPassive`, `IsChanneled`, `RefundsOnMiss` with `MissRefund()` (the 0.8 a
`RageCostOptions.Refund` takes), `PeriodicCanCrit`, `CanProcFromProcs`, `ClassSpellsOnly`,
`CannotCrit`, `IsAProc`, `SuppressesWeaponProcs` and `IsWeaponProcAura`; `IsBleed` reads
`SpellCategories.Mechanic` instead.

|                     |                                                                                                        |
| ------------------- | ------------------------------------------------------------------------------------------------------ |
| `effect.Trigger()`  | the spell `EffectTriggerSpell` names                                                                   |
| `spell.Triggered()` | every spell the row's effects fire, deduped, in effect order                                           |
| `spell.Drivers()`   | the spells whose effects fire this one, and the ones whose actionbar override replaces a spell with it |
| `spell.Refs()`      | the spells the tooltip names (`$12880d`, `$12966n`), in the order it names them                        |

An id the store does not carry is left out of those lists rather than answered as `Nil`. Lightning
Shield rank 1 shows why the two kinds of edge are separate: its aura effect triggers 26545, the
dispatcher every rank shares, while the rank's own damage spell 26364 is reachable only as a tooltip
reference.

`Spell.ClassFlags` is `SpellClassSet` with `SpellClassMask_0..3`: the family a spell files under and
its bit in it. `Effect.ClassFlags` is the same for `EffectSpellClassMask`, which is the set of spells a
modifier effect reaches, and `spell.AffectedBy(effect)` asks whether this spell is one of them.
`core.SpellConfig.ClassFlags` carries the row's own, which is what puts a registered spell inside the
talents that name it.

Class flags name the spells a modifier reaches, not one ability each: Mortal Strike and Bloodthirst
share word 0 `0x02000000`, and Retaliation's strike 20240 shares word 0 `0x80000000` with Whirlwind.
A handler or listener asking whether a spell is one ability takes `spellData.X.Contains(spell)`, which
matches the spell id against every rank, so `spellData.Execute.Contains(spell)` hears all five.

## A class file is ladders

Every class file is a generated `spellData` global with a `spelldata.Ladder` per family and no numbers
of its own (`sim/warrior/spell_data_auto_gen.go`):

```go
var spellData = generatedSpellData{
	AngerManagement: spelldata.Ranked(12296),
	Anticipation:    spelldata.Talent(12297, 5),
	BattleShout:     spelldata.Ranked(6673, 5242, 6192, 11549, 11550, 11551, 25289),
}
```

`Ranked` is one spell per rank, lowest first. `Talent` is a trait-tree talent, which the game files state
as one spell whose per-rank numbers live in a curve: each rank is that spell with the curve's value on
the effects the curve covers, and an effect it has no row for keeps the spell's own base value. Both
`MustFind` every id, so a family that loses a spell fails at startup rather than registering nothing.

|                                          |                                                                                          |
| ---------------------------------------- | ---------------------------------------------------------------------------------------- |
| `Rank(n)`                                | the spell at rank n. Rank 0 is untaken and answers `Nil`, so no `if rank > 0` guard      |
| `Highest()`                              | the top rank                                                                             |
| `ByID(id)`                               | the rank with this spell id; panics on one the ladder does not carry                     |
| `Len()`, `Each(fn)`                      | how many ranks, and each of them with its number                                         |
| `Contains(spell)`                        | whether a registered spell's id is one of the ladder's ranks                             |
| `ValueAt(rank)`                          | the rank's only effect in the game files' units; panics where the rank has more than one |
| `FractionAt`, `MultiplierAt`, `TenthsAt` | the same over 100, as `1 +` that, and over 10                                            |
| `EffectAt(n)`, `Effect(aura, misc)`      | one named effect across the ranks, carrying the same four readers                        |

`MultiplierAt` takes its sign from the data: a talent the game files state as -2/-4/-6 gives 0.94 at rank
3 and nobody writes the minus.
