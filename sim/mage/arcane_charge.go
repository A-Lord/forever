package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneBlastTriggered = spellData.ArcaneBlastTriggered.Highest()
var damageModIncrease = arcaneBlastTriggered.Effect(dbcenums.A_ADD_PCT_MODIFIER, 0).Percent()  // SPELLMOD_DAMAGE
var powerCostIncrease = arcaneBlastTriggered.Effect(dbcenums.A_ADD_PCT_MODIFIER, 14).Percent() // SPELLMOD_COST
var dotModIncrease = arcaneBlastTriggered.Effect(dbcenums.A_ADD_PCT_MODIFIER, 22).Percent()    // SPELLMOD_DOT

func (mage *Mage) registerArcaneCharges() {
	//damage mod only affects non-AB spells, and only direct damage (SPELLMOD_DAMAGE). Nothing
	//that lands after the cast takes it: charges gained meanwhile must not raise it. Those lock
	//the bonus in at the cast instead (arcaneChargesDotBonuses): Fireball's DoT its SPELLMOD_DOT,
	//the Arcane Missiles missiles their SPELLMOD_DAMAGE.
	damageMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll &^ (MageSpellArcaneBlast | MageSpellPyroblastDot | MageSpellFireballDot | MageSpellArcaneMissilesTick),
		FloatValue: damageModIncrease,
		Kind:       core.SpellMod_DamageDone_Pct,
	})

	//AB manacost mod
	abCostMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellArcaneBlast,
		FloatValue: powerCostIncrease,
		Kind:       core.SpellMod_PowerCost_Pct_Add,
	})

	config := spelldata.AuraConfig(arcaneBlastTriggered)

	config.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		damageMod.Activate()
		abCostMod.Activate()
	}

	config.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		abCostMod.Deactivate()
		damageMod.Deactivate()
	}

	config.OnStacksChange = func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
		stacks := float64(newStacks)
		damageMod.UpdateFloatValue(damageModIncrease * stacks)
		abCostMod.UpdateFloatValue(powerCostIncrease * stacks)
	}

	//New in Forever - AB Stacks modifies damage of all spells excluding AB, consuming the stacks
	config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if spell.Matches(MageSpellArcaneBlast) {
			return
		}

		aura.Deactivate(sim)
	}

	mage.ArcaneChargesAura = core.BlockPrepull(mage.RegisterAura(config))
}

// The charges' SPELLMOD_DAMAGE bonus as it stands, for a spell whose damage lands after the cast.
// Read it in ApplyEffects, before OnCastComplete consumes the charges.
func (mage *Mage) arcaneChargesDamageBonus() float64 {
	return damageModIncrease * float64(mage.ArcaneChargesAura.GetStacks())
}

// The charges' SPELLMOD_DOT bonus as it stands. Read it in ApplyEffects: the charges are consumed
// in OnCastComplete, before a travelling missile lands and applies its DoT.
func (mage *Mage) arcaneChargesDotBonus() float64 {
	return dotModIncrease * float64(mage.ArcaneChargesAura.GetStacks())
}

// The Arcane Charge DoT bonus each DoT locked in when it was applied (SPELLMOD_DOT names only
// Fireball and Frostfire Bolt). The ticks outlive the charges, so the bonus rides on the DoT's
// own PeriodicDamageMultiplier rather than on a spell mod.
//
// UNVERIFIED: whether the game applies the charges to these ticks at all; the row says so, the
// tooltip only says "the damage of all your other spells".
type arcaneChargesDotBonuses map[*core.Dot]float64

// lockIn sets the bonus a DoT carries until it is applied again: Fireball's DoT ticks read it
// through PeriodicDamageMultiplier, the Arcane Missiles missiles off their channel's DoT.
func (bonuses arcaneChargesDotBonuses) lockIn(dot *core.Dot, bonus float64) {
	dot.PeriodicDamageMultiplier *= (1 + bonus) / (1 + bonuses[dot])
	bonuses[dot] = bonus
}
