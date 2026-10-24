# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

## Weapon procs and extra attacks

- **A weapon proc whose row has melee or ranged DefenseType (direct hits and DoT-only applications) is a special hit:
  triggered by a main-hand or ranged hit it chains to weapon procs (its own weapon included), combat enchants and
  oils; triggered by an off-hand hit it is heard by none (main hand chains, off hand nothing). Magic weapon procs stay
  `ProcMaskEmpty`, equip and aura procs keep the `SpellFlagProc` gate, and Windfury never hears them.**
  `sim/common/shared/shared_utils.go`, `weaponProcStrikeMask`, PR #352.
- **The hit table decides a strike's mask, not its school (Bow of Searing Arrows is a ranged special); Shadow Oil
  hears its hand's melee hits and Frost Oil white hits only; a strike's own re-cast during its cast goes one batch
  later.** `sim/common/shared/shared_utils.go`, `procDamageHandler`, PR #352.
- **Extra attacks (every `ExtraAttackProc` user, Windfury Weapon) roll on the cast result and swing after the
  triggering hit (`sim.RunAfterCurrentEvent`), each swing only when it is in range (`InRange()`); Windfury Totem grants
  its buff row's charges (`ProcCharges`), a landed white hit spending one as it is dealt, without the attack power, and
  the extra swing the next; Heroic Strike may take the extra swing.** `sim/core/extra_attack.go`,
  `sim/core/buffs/drivers.go`, PRs #384 and #672.
- **Windfury Totem (10612) and Flametongue Totem (15036) are party Proc Trigger Spell auras that follow their rows
  (damage required, 100 ms recovery, main-hand mask kept), so Not-a-Proc hits such as Seal of Command trigger Windfury
  and Seal of Righteousness does not; ferals hear both, the imbue stays under either, Windfury Totem holds over
  Flametongue Totem, and Grace of Air Totem stacks with it.** `sim/core/buffs/drivers.go`, PR #158.

## Rates

- **A rate the game files do not state takes the carry-over value (`vanilla-carryover`) with a TODO, or is written and
  commented out waiting for a procs-per-minute rate, and the gear tab shows an in-game validation notice for every
  unvalidated rate, driven by the override tables' `Source`.** `tools/database/gen_effects.go`, `rateOrigin`, PR #322.
- **Omen of Clarity is 2.0 procs per minute with a 10 s ICD, turned into a chance per hit: a spell on its cast time
  after talent and set reductions (an instant 1.5 s, haste ignored), an auto attack or melee special on the main-hand
  swing speed. Only auto attacks and druid spells roll it, a periodic tick never does (Hurricane's ticks are instant
  hits in the sim and roll it under a TODO), and Moonkin Form doubles the chance and halves the ICD.**
  `sim/druid/omen_of_clarity.go`, `tools/database/overrides/spell_overrides.go`, PR #318.
- **Landed taunts do not roll Omen of Clarity (`ProcMaskEmpty`), and Faerie Fire rolls it only in caster form; read a
  proc's `ExtraCondition` before comparing spells' Omen behaviour.** `sim/druid/omen_of_clarity.go`, PR #651.
- **Revelation rolls on non-periodic damage and heals, including triggered-spell ticks such as Blizzard's and Vampiric
  Embrace heals (one heal event per party member), never on aura 3 periodic ticks; its rate is the chance the
  enchant's effect entry carries (`enchantProcChance`), with a TODO for the fall with the wearer's crit chance.**
  `sim/common/forever/enchants.go`, PR #246.

## Timing

- **A charge buff hit by the wearer's channel is reserved and spent at that channel's end (`Unit.OnChannelEnd`), an
  interrupted channel included, and a charge gained while a channel its class flags name is running is spent at that
  channel's end.** `sim/core/spelldata/charges.go`, `AttachChargeSpender`, PR #658.
- **Demonic Brand lands in the same batch as the Searing Pain hit that applies it (`TriggerImmediately`), as the combat
  log shows.** `sim/warlock/talents_demonology.go`, PR #414.
