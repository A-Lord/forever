package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var combustionRank = spellData.Combustion.Highest()
var combustionBuff = spellData.CombustionTriggered.Highest()

// The crit each stack grants, a percent on the buff row (10).
var combustionCritPerStack = combustionBuff.EffectN(1).BaseValue()

func (mage *Mage) registerCombustionSpell() {
	if !mage.Talents.Combustion {
		return
	}

	cd := core.Cooldown{
		Timer:    mage.NewTimer(),
		Duration: max(combustionRank.Cooldown(), combustionRank.CategoryCooldown()),
	}

	numCrits := 0

	critMod := mage.AddDynamicMod(core.SpellModConfig{
		ClassMask:  MageSpellFire,
		FloatValue: 0,
		Kind:       core.SpellMod_BonusCrit_Percent,
	})

	combustAura := mage.RegisterAura(core.Aura{
		Label:     "Combustion",
		ActionID:  core.ActionID{SpellID: combustionRank.ID},
		Duration:  core.NeverExpires,
		MaxStacks: int32(combustionBuff.MaxStack),

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			numCrits = 0
			critMod.Activate()
		},

		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			critMod.Deactivate()
			cd.Use(sim)
			mage.UpdateMajorCooldowns()
		},

		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks, newStacks int32) {
			critMod.UpdateFloatValue(combustionCritPerStack * float64(newStacks))
		},

		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !spell.SpellSchool.Matches(core.SpellSchoolFire) {
				return
			}

			if mage.Ignite != nil && spell.SameAction(mage.Ignite.ActionID) {
				return
			}

			if !result.Landed() {
				return
			}

			if numCrits >= 3 {
				aura.Deactivate(sim)
				return
			}

			aura.AddStack(sim)

			if result.DidCrit() {
				numCrits++
				if numCrits == 3 {
					aura.Deactivate(sim)
				}
			}
		},
	})

	config := spelldata.SpellConfig(&mage.Unit, combustionRank)
	config.Flags |= core.SpellFlagNoOnCastComplete | core.SpellFlagAPL
	config.ClassSpellMask = MageSpellCombustion
	// The cooldown starts when the buff fades, not when it is cast.
	config.Cast.CD = cd
	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !combustAura.IsActive()
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		combustAura.Activate(sim)
	}
	config.RelatedSelfBuff = combustAura

	combustSpell := mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: combustSpell,
		Type:  core.CooldownTypeDPS,
	})
}
