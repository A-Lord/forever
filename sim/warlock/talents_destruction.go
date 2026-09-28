package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (warlock *Warlock) registerDestructionTalents() {
	// Tier 1
	warlock.applyDestructiveReach()
	warlock.applyImprovedShadowBolt()
	warlock.applyBane()

	// Tier 2
	warlock.applyMoltenSkin()
	warlock.applyCataclysm()
	warlock.applyAftermath()

	// Tier 3
	warlock.applyRuin()
	warlock.applyShadowburn()

	// Tier 4
	warlock.applyIntensity()
	warlock.applyAgonizingFlames()
	warlock.applyConflagrate()

	// Tier 5
	warlock.applyPyroclasm()
	warlock.applyBaneOfHavoc()
	warlock.applyFireAndBrimstone()

	// Tier 6
	warlock.applyShadowAndFlame()

	// Tier 7
	warlock.applyIncinerate()
}

// --- Tier 1 ---

// TODO: Not implemented, as it does not affect sims for now: the row raises range and radius only.
func (warlock *Warlock) applyDestructiveReach() {
	if warlock.Talents.DestructiveReach == 0 {
		return
	}
}

// Shadow Bolt crits leave Shadow Vulnerability (17794) on the target: 12 sec of more Shadow damage taken from
// this warlock, with no charges, so another crit refreshes it. The row states rank 1; the talent prices the rank.
func (warlock *Warlock) applyImprovedShadowBolt() {
	if warlock.Talents.ImprovedShadowBolt == 0 {
		return
	}

	shadowVulnerability := spellData.ImprovedShadowBoltTriggered.ByID(17794)
	vulnerabilityMult := spellData.ImprovedShadowBolt.MultiplierAt(warlock.Talents.ImprovedShadowBolt)

	// Damage done by caster: this warlock's Shadow damage, against the target the debuff is on.
	warlock.ShadowVulnerabilityAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		auraConfig := spelldata.AuraConfig(shadowVulnerability)
		auraConfig.Label += " - " + warlock.Label
		return target.GetOrRegisterAura(auraConfig).AttachDDBC(DDBC_ShadowVulnerability, DDBC_Total, &warlock.AttackTables,
			func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
				if spell.SpellSchool.Matches(core.SpellSchoolShadow) {
					return vulnerabilityMult
				}
				return 1
			})
	})

	trigger := spelldata.ProcTrigger(&warlock.Character, spellData.ImprovedShadowBolt.Rank(warlock.Talents.ImprovedShadowBolt),
		func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			warlock.ShadowVulnerabilityAuras.Get(result.Target).Activate(sim)
		})
	// The talent row names no spells: the tooltip's are Shadow Bolt crits.
	trigger.ClassFlags = spellData.ShadowBolt.Highest().ClassFlags
	trigger.Outcome = core.OutcomeCrit
	warlock.MakeProcTriggerAura(trigger)
}

func (warlock *Warlock) applyBane() {
	if warlock.Talents.Bane == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Bane.Rank(warlock.Talents.Bane))
}

// --- Tier 2 ---

func (warlock *Warlock) applyMoltenSkin() {
	if warlock.Talents.MoltenSkin == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.MoltenSkin.Rank(warlock.Talents.MoltenSkin))
}

func (warlock *Warlock) applyCataclysm() {
	if warlock.Talents.Cataclysm == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Cataclysm.Rank(warlock.Talents.Cataclysm))
}

// Effect 1 is Conflagrate's chance to daze the target, which does not affect sims.
// TODO: The tooltip raises Immolate's initial damage, but effect 2 is SPELLMOD_DAMAGE, which the parse
// applies to all damage, the ticks included. Taken as all damage for now; test in-game, as it may be a bug.
func (warlock *Warlock) applyAftermath() {
	if warlock.Talents.Aftermath == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Aftermath.Rank(warlock.Talents.Aftermath), spelldata.Effects(2))
}

// --- Tier 3 ---

func (warlock *Warlock) applyRuin() {
	if warlock.Talents.Ruin == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.Ruin.Rank(warlock.Talents.Ruin))
}

func (warlock *Warlock) applyShadowburn() {
	if !warlock.Talents.Shadowburn {
		return
	}

	warlock.registerShadowBurn()
}

// --- Tier 4 ---

// TODO: Not implemented, as it does not affect sims for now: the row is pushback protection
// (SPELLMOD_NOT_LOSE_CASTING_TIME), which matters only once the sim models pushback from damage taken.
func (warlock *Warlock) applyIntensity() {
	if warlock.Talents.Intensity == 0 {
		return
	}
}

// TODO: Effect 2 is SPELLMOD_DAMAGE and effect 3 SPELLMOD_DOT on the same spells, so a dot they name takes
// both on its ticks. Taken as all damage for now; test in-game, as it may be a bug.
func (warlock *Warlock) applyAgonizingFlames() {
	if warlock.Talents.AgonizingFlames == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.AgonizingFlames.Rank(warlock.Talents.AgonizingFlames))
}

func (warlock *Warlock) applyConflagrate() {
	if !warlock.Talents.Conflagrate {
		return
	}

	warlock.registerConflagrate()
}

// --- Tier 5 ---

// TODO: Not implemented, as it does not affect sims for now: the row stuns (18093).
func (warlock *Warlock) applyPyroclasm() {
	if warlock.Talents.Pyroclasm == 0 {
		return
	}
}

func (warlock *Warlock) applyBaneOfHavoc() {
	if !warlock.Talents.BaneOfHavoc {
		return
	}

	warlock.registerBaneOfHavoc()
}

// Effect 1 is a dummy the tooltip does not use; effect 2 raises Conflagrate's crit.
func (warlock *Warlock) applyFireAndBrimstone() {
	if warlock.Talents.FireAndBrimstone == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.FireAndBrimstone.Rank(warlock.Talents.FireAndBrimstone))
}

// --- Tier 6 ---

// Conflagrate hits leave 10% more Shadow damage for 20 sec (1293816), Shadowburn hits 10% more Fire damage
// (426311), both parsed from their rows. Effect 2 is the chance Conflagrate keeps Immolate burning, 100% at
// 5/5; its chance to refund a Soul Shard is not modelled. The row's effects are dummies the tooltip reads.
func (warlock *Warlock) applyShadowAndFlame() {
	if warlock.Talents.ShadowAndFlame == 0 {
		return
	}

	rank := spellData.ShadowAndFlame.Rank(warlock.Talents.ShadowAndFlame)
	warlock.conflagrateKeepsImmolate = rank.EffectN(2).Percent()

	buffOnHit := func(buffRow *spelldata.Spell, hitBy spelldata.Ladder) {
		buff := warlock.RegisterAura(spelldata.AuraConfig(buffRow))
		spelldata.ParseEffects(&warlock.Character, buff, buffRow)

		warlock.MakeProcTriggerAura(core.ProcTrigger{
			Name:     buffRow.Name + " - Trigger",
			Callback: core.CallbackOnSpellHitDealt,
			Outcome:  core.OutcomeLanded,
			ExtraCondition: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
				return hitBy.Contains(spell)
			},
			Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
				buff.Activate(sim)
			},
		})
	}
	buffOnHit(spellData.ShadowAndFlameTriggered.ByID(1293816), spellData.Conflagrate)
	buffOnHit(spellData.ShadowAndFlameTriggered.ByID(426311), spellData.Shadowburn)
}

// --- Tier 7 ---

func (warlock *Warlock) applyIncinerate() {
	if !warlock.Talents.Incinerate {
		return
	}

	warlock.registerIncinerate()
}
