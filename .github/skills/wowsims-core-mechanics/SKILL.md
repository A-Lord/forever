---
name: wowsims-core-mechanics
description: "Use when WoWSims Forever Go code depends on what sim/core does around a spell that its row cannot show: the order of applyEffects, OnCastComplete and proc handlers (TriggerImmediately, SpellBatchWindow), costs, miss refunds, combo points and charges, how damage scales, how many targets an area hits, what CastRequirement refuses, registration hooks, whose auras a tag reads, aura lifetimes, exclusive categories, hit tables, resists, dot tick outcomes, pets, rage and threat, and proto field rules. Use it too when asked whether the sim models something (a racial, crit damage, resists, rage, stat scaling), when a handler sees or misses the cast that procced it, a refund goes wrong, another caster's dot counts, a registration panics, or a new proto bool reads false in saved settings. Reading a row is wowsims-spell-store; building an ability or aura is wowsims-abilities-auras; proc triggers are wowsims-procs; the pick at a cast end is wowsims-apl; runtime proofs are wowsims-testing; the map is wowsims-spells."
---

# wowsims-core-mechanics

A row says what a spell is; `sim/core` decides when its effects run, what the cast spends, which targets it reaches and
whose state it reads, so a spell that matches its row can still be wrong here. The rule this skill exists for: a claim
about core's timing (an order, a spend, a refund) is proven at runtime, never by reading the call order alone. The
rulings the code follows are in `references/rulings.md`; text here follows `wowsims-conventions`.

## Where things live

| Path                            | What it is                                                                   |
| ------------------------------- | ---------------------------------------------------------------------------- |
| `sim/core/cast.go`              | the cast: its checks, cost, GCD and cast time, then the effects              |
| `sim/core/spell.go`             | `RegisterSpell`, `CanCast`, `CanCompleteCast`, `applyEffects`, crit damage   |
| `sim/core/aura.go`              | aura registration and lifetime, tags, `NewAllyAuraArray`                     |
| `sim/core/aura_helpers.go`      | `core.ProcTrigger` and its deferral, `MakePermanent`                         |
| `sim/core/dot.go`               | `Dot.Apply`, ticks and refreshes                                             |
| `sim/core/spell_outcome.go`     | the outcome functions: hit, crit, block and the tick outcomes                |
| `sim/core/spell_result.go`      | `SpellChanceToMiss`, crit suppression, `DealCastResult`                      |
| `sim/core/spell_resistances.go` | partial resists and binary spells                                            |
| `sim/core/target.go`            | `NewAttackTable`: level difference, spell miss; `sim/core/attack.go` weapons |
| `sim/core/exclusive_effect.go`  | exclusive categories and effects                                             |
| `sim/core/energy.go`            | energy and combo points; `sim/core/rage.go` and `sim/core/mana.go` likewise  |
| `sim/core/pet.go`               | pets: enable and disable, stat inheritance, start distance                   |

| You are…                                                                     | Open                                                                        |
| ---------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| asking what the rotation sees at a hardcast end                              | `docs/rotation.md`, then `wowsims-apl`                                      |
| deciding a hit, crit, resist, cost, pet or threat number the row leaves open | `references/rulings.md`                                                     |
| chasing a trap: an item swap, weapon reach, `GainHealth`, pets, a proto bool | `references/traps.md`                                                       |
| looking for the review case behind a section below                           | `.github/skills/wowsims-review/references/defects.md`, the matching section |

## Cast order

- **`applyEffects` runs before `OnCastComplete`** (`cast.go`): `OnApplyEffects`, then the spell's `ApplyEffects`.
- **A proc handler runs inside `applyEffects` only with `TriggerImmediately`.** Without it, which is how
  `spelldata.ProcTrigger` builds one, it is a pending action one `SpellBatchWindow` (10 ms) later (`aura_helpers.go`),
  after `OnCastComplete`; a stepped test needs `sim.Step()` to see it fire.
- **`CallbackOnCastComplete` carries no result**, so `Outcome` on such a trigger is never read. A proc off the cast's
  hit roll uses `CallbackOnCastResult` (`references/rulings.md` §Cast order and procs). Prove an order at runtime:
  `wowsims-testing` `## Proving a claim at runtime`.

## What a cast spends

- **Core spends the cost at the cast; a refund is the spell's own call.** `spell.IssueRefund(sim)` gives back the cost's
  `Refund` share and a mana cost refunds nothing (`mana.go`). `SpellConfig` sets the share from the row's `MissRefund()`
  (0.8 where the row flags Discount Power On Miss), the default unless verified spell-specific behaviour differs.
- **Combo points are not part of the cost.** The spell spends them through `SpendComboPoints` (`energy.go`), each
  resource on its own: a failed Ferocious Bite keeps its points and loses its energy.
- **A one-charge buff is spent by the cast it modifies**; the spender is `wowsims-procs`. A spell's own charges
  (`MaxCharges`) are spent by `ConsumeCharge` and refuse the cast at zero.

## What scales

- **A flat weapon bonus is added before the weapon percent multiplies it** (`WeaponStrike` in `sim/core/spelldata/`).
- **`PointsPerResource` is the amount per combo point** (a parsed stat takes it at full points), and a tick that should
  scale with stacks or combo points reads them.
- **Crit damage** is `(base × CritMultiplierPct × CritDamageMultiplier × table − 1) × (CritMultiplierAdditive + 1) + 1`
  (`spell.go`); how aura 163 and additive talents meet is in `references/rulings.md` §Hit, crit and resist.

## How many targets

- **An area effect reaches every enemy in it** (`Effect.HitsAnArea()`, `HitsACone()`, `HitsAroundTheCaster()` in
  `sim/core/spelldata/resolve_spell.go`). `MaxTargets` has no `SpellConfig` field: a caller that caps reads
  `s.MaxTargets`, and "up to $s3 targets" or `SPELLMOD_JUMP_TARGETS` is the total count. Reach from a radius and
  party-aura reach are in `references/rulings.md` §Targets and reach.

## What the row forbids

- **Core checks `CastRequirement` on every cast**: form or stance, weapon or shield, "not usable in combat"
  (`castRequirementFailure` in `shapeshift.go`), in `Cast` and in `CanCompleteCast` (so in `CanCast`), before
  `ExtraCastCondition`, which never restates it (`wowsims-abilities-auras` §An ability; ranks: `auras-and-dots.md`).

## Hooks

- **A hook is attached once, at registration, on every path that needs it** (every attack that breaks stealth, a
  queued swing for every spec that registers it), and never in a reset, which runs every iteration. A rule every class
  needs goes in core once. `RegisterSpell` panics past 1000 spells, the sign of a per-iteration registration, and on a
  `DamageMultiplier` or `ThreatMultiplier` other than 0 with `ProcMaskUnknown` (`spell.go`).

## Whose state

- **An effect on several targets reads each hit target's state**, never the primary target's, and a map read for a key
  it may lack does not stand for 0.
- **Tags are every caster's.** `GetAurasWithTag` on a unit returns every registered aura with the tag, any caster's,
  active or not, and `HasAuraWithTag` is true for a registered aura that is not up (`aura.go`); filter by caster and use
  `GetActiveAuraWithTag` or `IsActive()`. A per-caster debuff's label and slots are `wowsims-abilities-auras` §An aura
  or dot.

## Auras and exclusive effects

- **`MakePermanent` sets `NeverExpires` and re-activates the aura in its `OnReset`** (`aura_helpers.go`); an aura with
  duration 0 panics on activation (`aura.go`).
- **A dot tick rolls its amount, not its hit.** `PeriodicDamageTick` deals one `Effect.Roll` per tick, and the default
  `TickOutcome` rolls no hit, since the application rolls it once (`sim/core/spelldata/resolve_aura.go`). A spell whose
  ticks roll hit passes `Spell.OutcomeTickMagicHitAndCrit` or `Dot.OutcomeTickMagicHit` (`spell_outcome.go`). Always
  Hit on a triggered row: `.github/skills/wowsims-abilities-auras/references/rulings.md`.
- **Within an exclusive category the enabled effect with the highest `Priority` applies**; `YieldsTies` takes the
  category only by outbidding (`exclusive_effect.go`). The `SingleAura` reset is `wowsims-abilities-auras` §Buffs and
  debuffs.

## Protos

- **On a merge, master keeps its field numbers and the branch renumbers its own new fields.** Nothing converts saved
  data (`ProtoVersion` in `proto/common.proto`), so renumbering a branch field costs nothing, while renumbering master's
  breaks every branch built on it.
- **No proto3 `optional` field**: the `npx protoc` behind `make ui/generated/proto/api.ts` rejects it. "Unset means X"
  is a plain field whose zero value means none or the default, or a message-typed field.
- **No `reserved` line before launch**, and a breaking change needs no migration; the full ruling is in
  `references/rulings.md` §Protos and saved data.

## Keeping this skill true

The code is right where this skill and the code disagree; fix the skill in the same change. Re-derive a claim before
repeating it:

- the cast order: `grep -n 'applyEffects\|OnCastComplete' sim/core/cast.go`, then
  `grep -n SpellBatchWindow sim/core/aura_helpers.go sim/core/constants.go`
- the requirement checks: `grep -n 'hasCastRequirement' sim/core/cast.go sim/core/spell.go`
- the refunds: `grep -n 'func (s \*Spell) MissRefund\|IssueRefund' sim/core/spelldata/attributes.go sim/core/*.go`
- the scaling: `grep -n 'return (flat\|CritMultiplierAdditive+1' sim/core/spelldata/spell.go sim/core/spell.go`
- the area reads: `grep -n 'MaxTargets has no\|func (e \*Effect) Hits' sim/core/spelldata/resolve_spell.go`
- the tick outcome: `grep -n 'func (s \*Spell) TickOutcome' -A9 sim/core/spelldata/spell.go`
- the tag reads: `grep -n 'func (at \*auraTracker) [A-Za-z]*WithTag' sim/core/aura.go`
- the panics: `grep -n 'Over 1000 registered\|ProcMask for spell\|Aura with 0 duration' sim/core/spell.go sim/core/aura.go`
- the proto rules: `grep -n -A8 'message ProtoVersion' proto/common.proto`; `grep -n '^\s*optional ' proto/*.proto`
  prints nothing

| You changed…                                                 | Update                                                                            |
| ------------------------------------------------------------ | --------------------------------------------------------------------------------- |
| the order in `cast.go` or `applyEffects`, the proc deferral  | "Cast order", `wowsims-procs` §Building a trigger from a row                      |
| `IssueRefund`, `MissRefund`, `WeaponStrike`, crit damage     | "What a cast spends", "What scales"                                               |
| the `CastRequirement` checks in `cast.go` or `spell.go`      | "What the row forbids", `docs/spell_data/auras-and-dots.md`                       |
| a tick outcome, `exclusive_effect.go`, tags, `MakePermanent` | "Whose state", "Auras and exclusive effects", `docs/spell_data/auras-and-dots.md` |
| `ProtoVersion`, the proto toolchain, the field-number rule   | "Protos"                                                                          |
| a trap's code, a ruling's code or its measured number        | `references/traps.md`, `references/rulings.md`                                    |

Then run the shared path check:

```
sh .github/skills/check_paths.sh .github/skills/wowsims-core-mechanics
```

A ruling, trap or workflow fact learned while using this skill is written here, or into `references/rulings.md`, in
the same change that uses it. A memory file is a pointer to this skill, not a second copy.
