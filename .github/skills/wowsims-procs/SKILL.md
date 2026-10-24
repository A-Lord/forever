---
name: wowsims-procs
description: 'Use when WoWSims Forever Go code builds or fixes a proc: calling spelldata.ProcTrigger or its ProcOpts (PPM, Chance, WeaponProc), reading ProcChanceSource or a ProcChance of 100 or 101, reading what DecodeProcTypeMask made of a row, porting a proc-chance talent, taking an ICD from the row, spending a charged buff, or wiring an item, enchant, set, extra-attack or weapon-totem proc. Use it too when a trigger never fires, a bleed tick procs nothing, a hand trigger fires on every hit or on a dodge, the build panics with "states no proc chance", Windfury or an oil hears the wrong hits, a charge outlives the cast that should spend it, or a measured rate disagrees with the tooltip. Reading a row is wowsims-spell-store; the buff a proc applies is wowsims-abilities-auras; when a handler runs is wowsims-core-mechanics; item effect routing is wowsims-consumables-items; the proc audit is wowsims-db-pipeline; tests that prove a rate are wowsims-testing; the map of spell skills is wowsims-spells.'
---

# wowsims-procs

A proc is a listener the row describes: which hits it hears, how often it fires and how long it rests.
`spelldata.ProcTrigger` builds all of that from the row, and the caller adds the name, the handler and only what the
game files leave out. The rule this skill exists for: the `ProcChance` column is not a chance until
`ProcChanceSource` says so, and a listener typed by hand around it fires on the wrong hits, or on every hit. The
rulings the code follows are in `references/rulings.md`. Text here follows `wowsims-conventions`; the page below holds
the detail.

## Where things live

| Path                                        | What it is                                                                     |
| ------------------------------------------- | ------------------------------------------------------------------------------ |
| `sim/core/spelldata/resolve_proc.go`        | `ProcTrigger`, its `ProcOpt`s, `StatedChance`                                  |
| `sim/core/spelldata/charges.go`             | `AttachChargeSpender`, the listener that spends a charged buff                 |
| `sim/core/spelldata/item_proc.go`           | `ItemProcUnsupportedWithRate`: whether an item's rows say enough to build      |
| `sim/core/proc_types.go`                    | `DecodeProcTypeMask` and the `ProcHint` bits                                   |
| `sim/core/aura_helpers.go`                  | `core.ProcTrigger`, `MakeProcTriggerAura`, the roll, the ICD, `OncePerCast`    |
| `sim/core/procs.go`                         | `DynamicProcManager` and the procs-per-minute managers                         |
| `sim/core/buffs/drivers.go`                 | the Windfury Totem and Flametongue Totem procs                                 |
| `sim/common/shared/`                        | the item, enchant and set proc builders (`NewSpellData*Proc`)                  |
| `tools/database/overrides/weapon_procs.go`  | a chance-on-hit weapon's procs per minute, which the game files do not carry   |
| `tools/database/overrides/enchant_procs.go` | a permanent enchant's rate where its spell states none or the 100/101 sentinel |

| You are…                                                                               | Open                                                                    |
| -------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| building a trigger from a row: the four chance shapes, what it fills, the rate options | `docs/spell_data/procs.md`, the opening tables                          |
| porting a proc-chance or proc-cooldown talent, a pet's or a target-side trigger        | `docs/spell_data/procs.md`, the paragraphs on `ProcSpellClassFlags`     |
| reading what the decoder made of a row, or a tooltip hint                              | `docs/spell_data/procs.md` §The decoder and the tooltip hints           |
| registering an item, enchant or set proc                                               | `docs/spell_data/procs.md` §Item and enchant procs                      |
| deciding what a named proc does, or which measured rate stands                         | `references/rulings.md`                                                 |
| reviewing a proc read off the wrong effect, a hand trigger, a charge nothing spends    | `wowsims-review`, `.github/skills/wowsims-review/references/defects.md` |

## Building a trigger from a row

What `ProcTrigger` fills and the Enrage example are in `procs.md`. What the page leaves to this skill:

- **Register it one of two ways.** `unit.MakeProcTriggerAura(trigger)` gives the listener an aura of its own, up all
  fight; `buff.AttachProcTriggerCallback(&unit, trigger)` hangs it on a buff, so it hears hits only while that buff is
  up. Either way `trigger.Name` is set first (see Traps).
- **The ICD is the row's** (`s.ICD()`); a typed one is a defect, as Hack and Slash's 500 ms against its row's 200 was.
- **A row with charges needs a spender.** `ProcCharges` becomes the buff's `MaxStacks` through `AuraConfig`, and
  nothing takes a stack off until `spelldata.AttachChargeSpender(character, buff, row, name)` is attached. It spends
  one charge per cast however many targets the cast hits (`core.OncePerCast`), and a channel's once at its end.
  Cutthroat 462707 had no spender, so one proc paid for every Ambush in its 10 s.
- **An extra attack goes through `unit.ExtraAttackProc(&trigger, tagSpellID, count)`.** It sets `TriggerImmediately`
  itself, swings once the triggering hit is dealt, never hears its own swing, and returns an `OnInit` for the proc's
  aura (`.ApplyOnInit`). The item shape is `shared.NewSpellDataExtraAttackProc`.
- **When a handler runs against the cast that procced it**, and what `TriggerImmediately` changes, is
  `wowsims-core-mechanics` `## Cast order`.

## Chance and masks

The four `ProcChanceSource` shapes, the decoder, the proc-chance talents and the `PPM`, `WeaponSpeedPPM`, `Chance` and
`ChanceFrom` options are in `procs.md`. Two options it does not list:

| Option              | What it does                                                                                                 |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| `WeaponProc()`      | the listener of a chance-on-hit effect or combat enchant: every landed weapon hit, the weapon deciding which |
| `ItemProcChance(s)` | an item or enchant proc rolls the column where it contradicts the effect's chance (Uther's Light 8397)       |

- **Options run in order, after the row.** Any option that reshapes the mask comes before the rate option, because
  `PPM` and the row's own procs per minute bind their manager to the mask they find. A `ProcMask` set on the returned
  trigger narrows the listener but not a manager already built.
- **A cast listener has no result.** On `CallbackOnCastComplete` the trigger's `Outcome` and `RequireDamageDealt` are
  never read and no manager rolls (`aura_helpers.go`). Initiative 13976 built that way awarded its point on a dodged
  Garrote; `ProcTrigger` on the rank decodes the hit listener and effect 1's 33/67/100%.
- **Read the chance off the effect the tooltip names**, never the one whose ladder fits: Hack and Slash's chance is
  effect 3, not effect 1's crit, though both hold 1 to 5.
- **A spell's own proc mask decides which listeners hear it**, and the hit table decides the mask, not the school.
  The cases (taunts on `ProcMaskEmpty`, a weapon proc's strike off either hand, oils) are in `references/rulings.md`.

## Item and enchant procs

Registration, the enchant slot rules, `ItemProcUnsupportedWithRate` and the two local snapshot lists are in `procs.md`
§Item and enchant procs; which builder an item effect routes to is `wowsims-consumables-items`.

- **More shapes than the page names.** Beside `NewSpellDataProc`, `NewSpellDataDamageProc` and `NewSpellDataAuraProc`,
  `sim/common/shared/` has `NewSpellDataHealProc`, `NewSpellDataAbsorbProc`, `NewSpellDataDebuffProc` and
  `NewSpellDataExtraAttackProc`.
- **A rate the game files lack is an override row**: `overrides.WeaponProcs` keyed by item and spell,
  `overrides.EnchantProcs` keyed by enchant and spell, each with `Reason` and `Source`. Rewriting the proc
  audit after new rows is `wowsims-db-pipeline`.

## Measuring a rate

- **Read the tooltip before the column.** `$h%` makes the column the chance; `$mN%` or `$sN%` makes that effect's
  ladder the chance; no chance and 100 or 101 means the proc fires on its own condition, and 101 usually means no
  proc at all; anything else is a procs-per-minute rate the game files lack (`fillProcChance` in `resolve_proc.go`).
  That rate is an override onto the row's `RPPM` first (`Field: PPM` in `spell_overrides.go`, Omen of Clarity is the
  case) or an override row on gear, and a `PPM(n)` literal with a review TODO only where no override fits.
- **An unstated rate takes the carry-over value with a TODO**, or is written and commented out until a rate exists.
  Its `Source` drives the gear tab's notice for an unvalidated rate (`rateOrigin` in `tools/database/gen_effects.go`,
  `ui/features/gear/item_notices.tsx`): `vanilla-carryover` and `tbc-carryover` name their origin, a measured rate's
  `wcl:`, `issue #` or `combat log, ` source shows none. A gear rate kept as a literal in the sim is listed in
  `literalRates` there, or the notice never names it.
- **The class `.results` never move on a proc rate.** Prove one with a `sim/common/shared` test that builds the
  listener and counts casts (`sim/common/shared/spell_data_weapon_rate_test.go`); see `wowsims-testing`
  `## Proving a claim at runtime`.
- **A measured rate that disagrees with the tooltip** is decided by `wowsims-spells`
  `## When a sim number disagrees with the game files`, and the outcome goes into `references/rulings.md` with its
  PR, as Omen of Clarity's and Revelation's did.

## Traps

- **A trigger named after its row never hears anything when the buff shares that name.** `ProcTrigger` names the
  trigger after the row, and `MakeProcTriggerAura` registers through `GetOrRegisterAura`, which hands back the
  `AuraConfig` buff of the same label and drops the listener. Set `trigger.Name = "<X> - Trigger"`.
- **A periodic listener built from a row never hears a bleed.** `DecodeProcTypeMask` files
  `PROC_FLAG_DEAL_HARMFUL_PERIODIC` under `ProcMaskSpellDamage`, and a bleed tick carries a melee mask. Set
  `trigger.ProcMask = core.ProcMaskUnknown` at the call site, as Thousand Cuts does (`sim/rogue/talents_subtlety.go`).
- **A harmful-spell-only mask (0x10000) decodes to a cast trigger that ignores misses.** The decoder keeps that rule
  for items, so a spell that must land to proc moves to the hit listener at its call site, as Nature's Grace does
  (`sim/druid/talents_balance.go`, PR #382).
- **A `core.ProcTrigger` typed by hand with no chance fires on every qualifying hit.** Core turns a `ProcChance` of 0
  into 1; only `spelldata.ProcTrigger` panics on a row that states no rate.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. Re-derive a claim before
repeating it:

- the trigger and its options: `grep -n 'func Proc\|) ProcOpt {' sim/core/spelldata/*.go`
- the decoder: `grep -n 'func DecodeProcTypeMask' sim/core/proc_types.go`
- the zero chance turned into 1: `grep -n 'ProcChance == 0' sim/core/aura_helpers.go`
- the item proc builders: `grep -n '^func NewSpellData[A-Za-z]*Proc' sim/common/shared/*.go`
- the bleed call sites: `grep -rn 'ProcMask = core.ProcMaskUnknown' sim/`

| You changed…                                                | Update                                                                       |
| ----------------------------------------------------------- | ---------------------------------------------------------------------------- |
| a `ProcOpt` in `resolve_proc.go`                            | "Chance and masks", the option table in `docs/spell_data/procs.md`           |
| `DecodeProcTypeMask` or a `ProcHint`                        | "Traps", `docs/spell_data/procs.md` §The decoder and the tooltip hints       |
| `AttachChargeSpender` or `ExtraAttackProc`                  | "Building a trigger from a row"                                              |
| a `NewSpellData*Proc` builder or an override table's fields | "Item and enchant procs", `docs/spell_data/procs.md` §Item and enchant procs |
| a ruling's code or its measured number                      | `references/rulings.md`                                                      |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-procs
```

A ruling, trap or workflow fact learned while using this skill is written here, or into `references/rulings.md`, in
the same change that uses it. A memory file is a pointer to this skill, not a second copy.
