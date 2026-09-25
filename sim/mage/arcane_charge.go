package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneBlastTriggered = spellData.ArcaneBlastTriggered.Highest()
var damageModIncrease = arcaneBlastTriggered.Effect(dbcenums.A_ADD_PCT_MODIFIER, 22).Percent()
var powerCostIncrease = arcaneBlastTriggered.Effect(dbcenums.A_ADD_PCT_MODIFIER, 14).Percent()

func (mage *Mage) registerArcaneCharges() {
	//damage mod only affects non-AB spells
	damageMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellsAll ^ MageSpellArcaneBlast,
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
