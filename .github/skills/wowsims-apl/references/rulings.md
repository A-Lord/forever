# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

## The pick at the cast end

- **At a player's hardcast end the rotation picks at the cast-end timestamp, before the cast's outcome applies, and a
  pushback re-queues the pick; the major cooldowns that go ahead of it and the units that look after the landing are
  `docs/rotation.md` §The flow.** `sim/core/gcd.go`, PR #640.
- **During the hardcast the pick reads the cast's own dot, own-id aura, cooldown and mana as landed; the full reading
  and the lines that need `Spell Is Casting` are `docs/rotation.md`.** `sim/core/apl_values_aura.go`, PR #640.
- **Include Reaction Time delays every aura change the same way, the unit's own casts included; during a hardcast the
  cast's own aura reads as landed instead.** `sim/core/apl_values_aura.go`, PR #638, PR #640.

## Values

- **Spell Expected DPET assumes every cast lands: the initial damage uses `OutcomeExpectedMagicCrit`, ticks the
  always-hit or crit outcome, divided by the longer of the hasted GCD and the cast or channel time; DoT clipping and
  mana are left out, and travel time only moves the landing.** `sim/core/apl_values_spell.go`, PR #580.
- **Spell Expected DPET counts only what lands before the fight ends (`sim.GetRemainingDuration()`): DoT ticks from the
  landing, channel ticks from the cast start, and a tick at exactly the fight end counts.**
  `sim/core/apl_values_spell.go`, PR #599.
- **Spell Expected DPET uses the average partial resist instead of rolling one.** `sim/core/spell_resistances.go`,
  PR #644.
- **`Is Hit Capped` and `Is Expertise Capped` are per spell, current target only and yes or no: melee caps at the
  special-attack soft cap without the dual-wield penalty, the table follows the spell's `DefenseType` (proc mask for
  none), auto shots roll the ranged table, and the 1e-9 tolerance stays in the APL values.**
  `sim/core/apl_values_spell.go`, PR #608.
- **Bane Is Active On All Targets and Assigned Curse Is Active On All Targets are true when every active target holds
  one, read as raw `IsActive()`; with NoCurse the curse value reads false without a validation warning.**
  `sim/warlock/apl_values.go`, PR #654.
- **`Multiple CD Usages` with no base spell aligns casts to the encounter end early only when a later use still gets at
  least half its buff (remaining > CD + buff/2) or the buff runs to the end.** `sim/core/apl_values_misc.go`, PR #610.
- **An empty base spell in `Multiple CD Usages` means none, and a potion's cooldown there is `Spell.FullCooldown()`, the
  larger of its cooldown and shared cooldown.** `sim/core/apl_values_misc.go`, PR #625.

## Actions and validations

- **A cast whose spell never reaches the starting distance warns, unless the rotation has a Move, a custom rotation, a
  `SpellFlagMovesToTarget` cast (Charge, Intercept) or a range mod not yet active.** `sim/core/apl_actions_casting.go`,
  PR #577.
- **Strict Multidot re-checks its entry's If and the If of every Group Reference above it before each cast after the
  first; Sequence and Schedule wrappers' Ifs only decide the start, and a gated next cast is never queued.**
  `sim/core/apl_actions_casting.go`, PR #603.
- **The Pet Attack prepull action (hunter and warlock) starts the pets' rotation and swings from its time, even before
  0, with damage before 0 uncounted; the destro and demo warlock builds use it at -1 s.** `sim/core/pet.go`, PR #578.

## Warrior presets

- **Fury and Arms align Death Wish, Recklessness, racials and the offensive potion through `Multiple CD Usages`, while
  the rage potion and Protection's Death Wish stay on cooldown.** `ui/specs/warrior/`, PR #610.
- **With the Recklessness group on, the warrior APLs cast Elune's Light only while Recklessness is inactive, more than
  15 s remain, and either more than 195 s or at most 30 s remain; the Recklessness line itself is untouched.**
  `ui/specs/warrior/dps/apls/`, PR #653.
- **The "Hamstring Weave ON/OFF" group (`use_hamstring`, on by default) weaves Hamstring for Windfury procs only while
  the totem is known (`auraIsKnown 10612` tag -1): Fury below Overpower weaving, Arms above Execute while the Windfury
  buff is inactive.** `ui/specs/warrior/dps/apls/`, PR #243.
- **The Sunder option: Help (default) stacks to 5 and leaves the rest to the raid, restacking after the debuff falls off
  even with no raid Sunder; Maintain stacks and refreshes; None casts none; the stack and refresh branches stay separate
  groups.** `ui/specs/warrior/dps/spec.ts`, PR #234.
