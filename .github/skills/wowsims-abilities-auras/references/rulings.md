# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

## Rolls and outcomes

- **`ATTR_EX_3_ALWAYS_HIT` on a triggered row stops only that row's own roll; the sim skips the miss only when the row
  the player casts carries the bit, so Judgement of Command keeps the miss roll of its dummy 20968.**
  `sim/core/spelldata/attributes.go`, `AlwaysHits`.
- **A spell whose row has DefenseType magic and no Always Hit rolls `OutcomeMagicHit` even when it deals no damage
  (Taunt, Challenging Shout, Intimidating Shout, Growl, Challenging Roar); taunts carry `ProcMaskEmpty`.**
  `sim/warrior/taunt.go`, PR #649.
- **Intercept's stun is its own spell from the stun row, rolling `OutcomeMagicHitAndCrit` with an empty proc mask;
  targets at character level + 3 are immune to its stun and damage, and nothing is dealt at fight-end teardown.**
  `sim/warrior/intercept.go`, PR #652.
- **Consecration ticks every second from a tick one second after the cast, each tick rolling its own hit, and Serpent
  Sting is a full ranged ability that reaches Judgement of Wisdom and ranged procs.** `sim/paladin/consecration.go`,
  `sim/hunter/serpent_sting.go`, PR #452.
- **DoTs do not snapshot: every tick deals damage on the stats and modifiers of the moment it lands
  (`CalcAndDealPeriodicDamage` in `OnTick`, no `OnSnapshot`), Deep Wounds' rollover pool being the one amount fixed at
  the crit.** `sim/core/spelldata/resolve_aura.go`, `DotConfig`.
- **Deep Wounds bleeds the talent share of average weapon damage without attack power; a new crit rolls the owed damage
  over and extends the timer without resetting it, and the ticks never crit.** `sim/warrior/talents_arms.go`, PR #231.

## Reading the row

- **When a row's effect type and its tooltip disagree, the effect type wins (Unleashed Fury 19616 is a flat aura 107
  modifier); a tooltip-only number is flagged, not adopted.** `sim/core/spelldata/parse_effects_table.go`, PR #468.
- **Two aura 108 rows multiply for cost (misc 14, `SpellMod_PowerCost_Pct`), direct damage (misc 0,
  `SpellMod_DirectDamageDone_Pct`) and periodic damage (misc 22); a flat 107 modifier applies before a 108 percent.**
  `sim/core/spelldata/parse_effects_table.go`, PR #378.
- **A talent's trait curve applies with the sign the game files state, even where base points and tooltip say the
  opposite; no sign override exists.** `sim/core/spelldata/ladder.go`, PR #399.
- **Lava Burst's Flame Shock bonus is 20% at every rank; a higher-rank row value is a tooltip bug.**
  `sim/shaman/lava_burst.go` reads `EffectN(2).Average()`, which gives 53% at rank 3 and disagrees with this rule;
  PR #487.
- **Seal of Righteousness keeps the tooltip description-variable coefficients 0.12 and 0.0765 over the rows' 0.**
  `sim/paladin/seal_of_righteousness.go`, PR #469.
- **Revenge deals its base plus 0.28 × attack power on every rank (`APCoefDirect`), measured on ranks 1 and 2 and
  carried to the rest with a TODO per row.** `tools/database/overrides/spell_overrides.go`, PR #256.
- **A weapon requirement on aura 52 or 54 is per weapon; on aura 290 with a melee-weapon mask it applies while equipped
  to melee, ranged and spells, with a ranged-only mask per weapon; tooltips saying "with <weapons>" (Weaponmaster
  12700, Hack and Slash 13706) are per weapon and hand-built; weapon-scoped spell mods skip magic-defense spells.**
  `sim/core/spelldata/weapon_scope.go`, PR #87.
- **Shield Bash and Shield Slam register as `ProcMaskMeleeOHSpecial`, which keeps Windfury off them; the game files give
  no hand, so this is a ruling, not data.** `sim/warrior/shield_bash.go`.

## Registering

- **Casters register every rank (melee registers `Highest()` only), each per-rank DoT or HoT labelled with
  `spelldata.RankLabel(rank, &owner.Unit)` so ranks never share one aura; category cooldowns already share one timer
  across ranks, so no hand `NewTimer()`.** `sim/core/spelldata/resolve_aura.go`, PR #246.
- **Devouring Plague is learnable by every race, and Inner Focus restarts its cooldown when the buff is spent
  (`OnExpire` calls `CD.Use`) though the row has no cooldown-on-event bit.** `sim/priest/devouring_plague.go`,
  `sim/priest/talents_discipline.go`, PR #246.
- **A school hit talent (SPELLMOD_RESIST_MISS_CHANCE on class flags) shows on the sheet by opting in with
  `spelldata.ShowAsSchoolHit(schools)` and its tooltip's schools, never `SchoolBonusHitChance`; all-spell hit talents
  (aura 55 misc 0) already show and must not opt in.** `sim/core/spelldata/parse_effects.go`, PR #285.
- **Weaponmaster covers a mace or staff in either hand without doubling, Anger Management comes from its aura, crit
  rage ×2 follows the row, a wand keeps `SpellFlagNoPartialResist`, and Flamestrike's burn stays a dot.**
  `sim/warrior/talents_arms.go`, `sim/mage/flamestrike.go`, PR #494.
- **Eureka! is an 8 s buff with no charges: `applyEureka` keeps the cost and direct-damage mods up for the whole buff
  and gives SPELLMOD_DOT channels their tick bonus only while it is up; a row with an empty SPELLMOD_DOT mask is
  skipped.** `sim/core/racials.go`, PR #462.
- **Cleave's flat threat splits by the targets it landed on, Revenge's `ThreatMultiplier` is 2.25, and Charge's rage
  carries `NoThreat` from `ATTR_EX_1_NO_THREAT`.** `sim/warrior/heroic_strike_cleave.go`, `sim/warrior/revenge.go`,
  `sim/warrior/charge.go`, PR #390.
- **The warlock spends only Soul Shard 6265 reagents, and Decimation's `A_NO_REAGENT_USE` effect waives the shard.**
  `sim/warlock/soul_shards.go`, PR #331.

## Forms

- **A shift into Bear or Dire Bear Form empties the rage bar in the shift spell's `ApplyEffects`, Furor's 10 rage comes
  after, and the starting form at the pull keeps `StartingRage`.** `sim/druid/forms.go`, PR #149.
- **A druid's change of form keeps the current health percentage both ways and a Dire Bear guardian starts the pull at
  100%; any new form or max-health form effect sits inside `keepHealthPercentAcrossForms`.** `sim/druid/forms.go`,
  PR #149.

## Buffs and debuffs

- **Every debuff lands and only the strongest in its category counts (`SingleAura: false`); buff categories keep the
  single-aura rule; Curse of Recklessness, Faerie Fire and Dust Cloud share `MinorArmorReduction`.**
  `sim/core/buffs/meta.go`, PR #159.
- **Dark Inspiration reaches party players only (`Character.NewPartyAuraArray`), bids in the party buff's category
  through `buffs.NewSpellDataDarkInspirationAuraFor`, yielding ties, and never buffs pets.** `sim/core/buffs/buffs.go`,
  PR #505.
