# Traps in core

Each line is a trap of `sim/core` that a spell's row cannot warn about, under the section of `SKILL.md` it belongs to.

## What a cast spends

- **An item swap that moves a weapon slot extends the GCD to at least now + 1.5 s and restarts both swing timers**
  (`sim/core/item_swaps.go`, `SwapItems`). The cost is not measured in game, and an APL with swaps needs
  `enableItemSwap: true`.

## How many targets

- **A `core.Weapon{}` literal leaves `MaxRange` 0, which `IsInRange()` reads as unlimited** (`sim/core/attack.go`), so a
  pet or form weapon that can stand past melee range sets `MaxRange: core.MaxMeleeRange` (PR #578).

## Hooks

- **An aura's `OnExpire` also runs at iteration teardown** (`doneIteration` in `sim/core/aura.go`), so arrival logic
  hangs on `MovementEnd`, not on an expiry.
- **Shared code hooks a pet through `Pet.RegisterEnableEffect` and `RegisterDisableEffect`**, since `OnPetEnable` is
  one class-owned field (`sim/core/pet.go`).
- **`GainHealth` fires no heal callback** (`sim/core/health.go`), so a heal that must roll procs (Vampiric Embrace) is
  a real heal spell (PR #246).
- **A bare stepped sim with no APL never swings** until `AutoAttacks.EnableMeleeSwing(sim)` (`sim/core/attack.go`).
- **A spec's own metric row goes through `spell.RecordCustomMetric(target, key, value)`** (label
  `results_tab.details.custom_metrics.<key>`), and a shield records through `spell.RecordShielding`; `Shield.Apply` has
  no callers (`sim/core/spell.go`, PR #510).

## Auras and exclusive effects

- **`Pet.Disable` expires only active auras**, which never-expiring auras never join, so a dismiss leaves permanent pet
  auras up (`sim/core/pet.go`).

## Protos

- **A new proto3 bool decodes as false from saved settings**, so a default-on toggle shows off for returning users
  (PR #243).
