package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The debuff half is A_MOD_SPELL_DAMAGE_FROM_CASTER: while the channel is on the target, it takes 10%
// more damage from the caster's spells the effect's class flags name. Those are Corruption and Bane of
// Agony (0x2 | 0x400); Siphon Life, a leech rather than periodic damage, and the drains are not named.
func (warlock *Warlock) registerWrackSpell() {
	rankedSpell := spellData.Wrack.ByID(1316697)
	wrackTick := rankedSpell.PeriodicEffect()
	wrackAmplify := rankedSpell.Effect(dbcenums.A_MOD_SPELL_DAMAGE_FROM_CASTER, 0)

	soulSiphonMult := warlock.soulSiphonPerEffect()
	ampMult := 1 + wrackAmplify.Percent()

	// Damage done by caster: this warlock's damage from the named spells, against the target Wrack is on.
	amplifyNamed := func(_ *core.Simulation, spell *core.Spell, _ *core.AttackTable) float64 {
		if spell.MatchesFlags(wrackAmplify.ClassFlags) {
			return ampMult
		}
		return 1
	}

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellWrack

	// TODO: Confirm no snapshot from soul siphon
	config.Dot = spelldata.DotConfig(rankedSpell, wrackTick)
	config.Dot.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		result := dot.Spell.CalcPeriodicDamage(sim, target, wrackTick.Average(dot.Spell.Unit.Level), rankedSpell.TickOutcomeHitRolled(dot))
		scaleResult(sim, dot.Spell, result, warlock.soulSiphonMultiplier(target, soulSiphonMult))
		dot.Spell.DealPeriodicDamage(sim, result)
	}

	// Uses attach DDBC methods to apply/remove the 10% damage taken buff.
	// Reference MoP Haunt Implementation for example
	config.Dot.Aura.OnGain = func(aura *core.Aura, _ *core.Simulation) {
		core.EnableDamageDoneByCaster(DDBC_Wrack, DDBC_Total, warlock.AttackTables[aura.Unit.UnitIndex], amplifyNamed)
	}
	config.Dot.Aura.OnExpire = func(aura *core.Aura, _ *core.Simulation) {
		core.DisableDamageDoneByCaster(DDBC_Wrack, warlock.AttackTables[aura.Unit.UnitIndex])
	}

	// Roll to hit and apply as dot
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHitNoHitCounter)
		if result.Landed() {
			spell.Dot(target).Apply(sim)
		}
		spell.DealOutcome(sim, result)
	}

	// Calculate expected damage, scaled by soul siphon multiplier.
	config.ExpectedTickDamage = func(sim *core.Simulation, target *core.Unit, spell *core.Spell, _ bool) *core.SpellResult {
		result := spell.CalcPeriodicDamage(sim, target, wrackTick.Average(spell.Unit.Level), spell.OutcomeExpectedMagicAlwaysHit)
		result.Damage *= warlock.soulSiphonMultiplier(target, soulSiphonMult)
		return result
	}

	warlock.Wrack = warlock.RegisterSpell(config)
}
