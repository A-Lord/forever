# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

## Cast order and procs

- **An APL aura condition reads an aura only once it is active: a ProcTrigger declares no pending aura, and a proc
  that must land in the same batch as its hit sets `TriggerImmediately`, only with log proof.**
  `sim/core/aura_helpers.go`, PR #638.
- **A proc that comes off the cast's hit roll uses the cast result hook (`CallbackOnCastResult` with
  `TriggerImmediately`, `Aura.OnCastResult`); `spell.DealCastResult` runs once per result, at the top of damage dealing
  for non-missile results and through `DealDamageAfterTravel` for missiles, never for ticks or heals.**
  `sim/core/spell_result.go`, PR #384.
- **Eclipse is gained with the Wrath's cast result when its hit roll lands, with no batch delay, and a missed or fully
  resisted Wrath grants none; Nature's Grace is gained at the crit's impact; a missed Starfire still spends a
  charge.** `sim/druid/talents_balance.go`, PR #382.
- **Nature's Grace is a 3 s buff of +10% spell haste and −10% GCD from non-periodic spell crits, never spent,
  activated one spell batch after the crit (default ProcTrigger deferral, no `TriggerImmediately`), and the GCD under
  it keeps both reductions (1.227 s).** `sim/druid/talents_balance.go`, PR #413.
- **Without Frostbite, Fingers of Frost rolls its own chance at the impact of each landed Chill hit and grants at once;
  with Frostbite it comes with Frostbite's rank chance at the Frostbolt cast plus an independent impact roll of
  (p − f)/(1 − f), so the total stays 15% at every rank, and a missed or fully resisted Frostbolt grants none.**
  `sim/mage/talents_frost.go`, PR #430.
- **Fingers of Frost is spent by every damaging rotation cast except channels and wand Shoot, never by DoT or burn
  ticks, and by each landed Arcane Missiles missile, a miss keeping the charge.** `sim/mage/talents_frost.go`, PR #343.
- **Every Fingers of Frost spend goes one `core.SpellBatchWindow` after the cast that spends it, from the pending
  action pool, so a cast queued in the same batch is still Frozen; the cast that granted the charges spends none, and a
  grant inside that batch leaves the charges full.** `sim/mage/talents_frost.go`, `SpellBatchWindow`.

## Hit, crit and resist

- **The combat model is level 60 with target levels 58 to 63: attack tables come from the level difference, capped at
  +3, and per-hand weapon skill.** `sim/core/target.go`, PR #70.
- **Weapon skill eligibility is per class, race only setting which skills are known at the start.**
  `sim/core/character_constants.go`, PR #200.
- **Spell hit can take the spell miss chance to 0; there is no 1% minimum (`SpellChanceToMiss`).**
  `sim/core/spell_result.go`, PR #200.
- **Spell miss against a target below the attacker is 4, 3, 2, 1 and 0% with no floor, and the +3 cap stays above.**
  `sim/core/target.go`, `spellMissChance`, PR #452.
- **Against a target three or more levels above, physical crit loses the skill part (0.2% per defense point) plus
  min(aura crit, 1.8%), aura crit being total crit minus class base crit minus Agility crit; against +1 and +2 only the
  skill part applies.** `sim/core/spell_result.go`, `sim/core/target.go`, PR #202.
- **Weapon-damage specials (effects 17, 31, 58, 121) and white hits never crit when blocked (block before crit);
  school-damage specials (effect 2) roll block and crit separately; pick the outcome function by the damage effect
  type.** `sim/core/spell_outcome.go`, PR #200.
- **Aura 163 (and aura 50 for healing) +X% multiplies `PseudoStats.CritDamageMultiplier`, which scales the base crit
  multiplier, so 2% gives 204% melee and 153% spell crits; additive crit-bonus talents then scale the bonus above 100%,
  so Vengeance 5/5 with 2% gives (1.53 − 1) × 2 + 1 = 206%.**
  `sim/core/spell.go`, PR #462.
- **`EffectVariance` is the full roll width: spell hits and heals roll the whole amount, power included, by ±V/2
  (`Spell.DirectVariance` after power), and physical effects roll the base by the full V before attack power.**
  `sim/core/spell_result.go`, PR #267.
- **Every DoT tick rolls its effect's amount.** `sim/core/spelldata/resolve_aura.go`, `PeriodicDamageTick`, PR #573.
- **Tick misses count in `TargetedActionMetrics.tick_misses`, not `Misses`, and Consecration's own cast roll does not
  count (`OutcomeTickMagicHitNoHitCounter`).** `sim/core/spell_outcome.go`, PR #570.
- **Partial resists roll the 10% bucket triangle with a = R/(abs(R) + 9 × attacker level); Spell Piercing beyond the
  target's resistance is bonus damage, the target's own resistance floors at 0, and there is no level-based resist.**
  `sim/core/spell_resistances.go`, PR #436.
- **DoT ticks resist at a tenth of the resistance by default; channels and spells with `SpellFlagFullResistanceTicks`
  take it in full.** `sim/core/spell_resistances.go`, PR #288.
- **A landed binary spell rolls the partial and piercing roll as well, and damage shields carry
  `SpellFlagNoPartialResist`.** `sim/core/spell_resistances.go`, PR #420.
- **A wand shot misses on the spell table at wand skill (capped at +3), crits on ranged crit with skill suppression for
  ×1.5, takes no partial resist, dodge, parry or block, carries the ranged-auto proc mask and the wand's school, and
  adds no attack or spell power.** `sim/core/wand.go`, PR #249.

## Stats

- **Aura 674 armor is its own stat `ArmorNoMods`, added to Armor through one unmultiplied stat dependency, so percent
  armor modifiers and equip scaling never touch it.** `sim/core/unit.go`, PR #471.
- **No game file ships class base crit; the hand-kept tables in `assets/db_inputs/basestats/` hold it (melee: paladin
  0.7, priest 3.0, shaman 1.7, mage 3.2, warlock 2.0, druid 0.9, others 0; spell: paladin 3.5, hunter 3.6, priest 0.8,
  shaman 2.3, mage 0.2, warlock 1.7, druid 1.8), and Agility to crit comes from `PlayerExpectedStat`.**
  `assets/db_inputs/basestats/`, PR #203.
- **Every race but Skyborne takes the stat offsets in `RaceOffsets`, Spirit included; the Skyborne rows keep their
  sheet values until fresh level 1 naked sheets show their own offsets.** `sim/core/base_stats.go`, PR #619.
- **Creature-type racial slaying (aura 168: Troll Beast Slaying, the Dwarf and Skyborne rows) raises damage only
  (`applyDamageVersus`); crit damage against a creature type comes only from aura 547 rows.** `sim/core/racials.go`,
  PR #536.
- **Touch of the Grave takes no downrank scaling from a low-level trigger and no Shadow modifiers; its amount and 25 s
  ICD come from the row.** `sim/core/racials.go`, PR #462.

## Costs, resources and threat

- **Mana regenerates in 100 ms ticks; a paid channel holds regen until max(start + 5 s, end) and a free channel does
  not.** `sim/core/mana.go`, PR #249.
- **OOM runs while the highest-priority spell the APL wants (condition true, ready but for mana) is skipped for mana,
  whatever casts instead; it starts at the first free rotation pass (GCD ready, no hardcast) that skips one and ends at
  the first that skips none.** `sim/core/spell.go`, `meetsCost`, PR #371.
- **Rage from damage taken is 2.5 × (1 − A) × pre-mitigation damage / c, with A by attacker level (20% at 42 and
  below, 30% at 44, 36% at 46, 38% at 50, 40% at 63, interpolated between rows) and c the rage conversion; defender
  level, health and armor do not matter.** `sim/core/rage.go`, PR #356.
- **Rage threat is booked in `AddRage` at gain time, 5 × the unit's threat multiplier split across the active targets,
  and rage from white hits, hits taken and refunds carries `ResourceMetrics.NoThreat`; a new rage source that comes
  from damage sets `NoThreat`.** `sim/core/rage.go`, PR #385.
- **Mana gained makes `ThreatPerManaGained` threat split between the targets in combat at the time; regen and sources
  marked `NoThreat` make none.** `sim/core/mana.go`, `addManaGainThreat`.
- **A healing-only on-use buff is `CooldownTypeHealing`, a Stamina-only one is survival, and Admiral's Hat is
  survival.** `sim/core/aura_helpers.go`, `CooldownTypeOf`, PR #614.

## Targets and reach

- **A castable row centred on the caster or a cone takes its radius as MaxRange, shouts with no damage included;
  procs, ticks and totem pulses stay ungated, combat reach is not added, and a periodic-trigger aura is followed one
  hop.** `sim/core/spelldata/resolve_spell.go`, PR #577.
- **In-sim casts of party or raid area auras (paladin auras, shaman totems) reach only the caster, the party and raid
  buff settings standing in for the others; raid-targeted buffs (Battle Shout, the prayers) use `NewAllyAuraArray`.**
  `sim/core/aura.go`, PR #505.
- **Wand Shoot is a channel aura that never ticks plus shots from a wand-owned slot: wand speed is hasted like a cast
  time and fixed at the start, the first shot comes at max(cast + 0.5 s wind-up, last real cast + speed) and then every
  speed, and it stops only for a ready or queueable spell at a slot or for movement.** `sim/core/wand.go`, PR #249.

## Pets

- **Demons take the warlock's spell crit as both melee and spell crit, health 7.5 per warlock Stamina on top of their
  own, 35% of the warlock's armor (Demon Armor included), and 17% (attack power) and 10% (spell damage) of the
  warlock's spell damage; their own level-60 stats are shares of ExpectedStat.** `sim/warlock/pets.go`,
  `petStatInheritance`.
- **Dynamic pets (demons, hawks) refresh their inherited stats from the owner's totals at the pet heartbeat instead of
  running stat deltas through the inheritance.** `sim/core/pet.go`, PR #432.
- **Hawks take all ranged-only haste at once through `PetConfig.HasDynamicRangedSpeedInheritance` (not at the
  heartbeat), with a 2.5 s base swing.** `sim/hunter/summon_hawk.go`, PR #452.
- **The Felhunter's mana regen is 10% of the warlock's Intellect, the Voidwalker's and Sayaad's share (the Imp's is
  30%), until it is read in game.** `sim/warlock/pets.go`, PR #545.
- **Every demon starts at the warlock's distance (`StartsAtOwnerDistance`), a mid-fight summon appears there, melee
  pets walk to `MaxMeleeRange − 1` before acting, and the Imp stays put unless the warlock stands past Firebolt's
  30 yd; warlock suites run at `StartingDistance: 20`.** `sim/core/pet.go`, `sim/warlock/pets.go`, PR #578.
- **True pets stand under every totem the owner does (the same constructor on the pet unit, toggled with the owner's
  copy and the pet's enable and disable), unless the aura row carries `ATTR_EX_3_ONLY_ON_PLAYER`
  (`Spell.OnlyOnPlayers`); guardians such as Shadowfiend get none, and Windfury and Flametongue roll on the pet's own
  melee.** `sim/core/buffs/totems.go`, PR #370.

## Metrics and encounters

- **Per-hit ranges in action metrics follow their totals: damage ranges record in combat only, healing and shield
  ranges always; ranges live on the Spell and merge by Welford and Chan, so single- and multi-thread results match.**
  `sim/core/spell.go`, `sim/core/utils.go`, PR #510.
- **The generic target presets use 3731 armor and 1 428 900 health.** `sim/encounters/register_all.go`, PR #442.
- **Generic targets hit with MinBaseDamage 2601 (Chromaggus), and the protection warrior's default healing model is
  2.4 × the Defense Cap set's DTPS on that target with cadence 0.3 s and variation 0.64.**
  `sim/encounters/register_all.go`, `ui/specs/warrior/protection/presets.ts`, PR #316.

## Protos and saved data

- **The sim is not live, so a change that breaks saved APLs or settings needs no migration, only a note in the PR body;
  a removed proto field leaves no `reserved` line, new fields refill freed numbers and every `// NextIndex:` is the
  highest used number + 1.** `proto/common.proto`, PR #149.
