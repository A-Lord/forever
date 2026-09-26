package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var fireWardRank = spellData.FireWard.Highest()
var frostWardRank = spellData.FrostWard.Highest()
var manaShieldRank = spellData.ManaShield.Highest()
var iceBarrierRank = spellData.IceBarrier.Highest()
var iceBlockRank = spellData.IceBlock.Highest()
var counterspellRank = spellData.Counterspell.Highest()

func (mage *Mage) registerDefensiveSpells() {
	mage.registerAbsorbSpell(fireWardRank, MageSpellFireWard, core.SpellSchoolFire)
	mage.registerAbsorbSpell(frostWardRank, MageSpellFrostWard, core.SpellSchoolFrost)
	// Mana Shield burns mana per point absorbed; the drain is not modelled, only the absorb.
	mage.registerAbsorbSpell(manaShieldRank, MageSpellManaShield, core.SpellSchoolNone)
	mage.registerIceBarrier()
	mage.registerIceBlock()
	mage.registerCounterspell()
}

// A ward: an absorb shield whose strength is the row's first effect, against the school its
// misc value names (none absorbs everything).
func (mage *Mage) registerAbsorbSpell(rank *spelldata.Spell, mask int64, school core.SpellSchool) {
	absorbAura := mage.NewDamageAbsorptionAura(core.AbsorptionAuraConfig{
		Aura: core.Aura{
			Label:    rank.Name + mage.Label,
			ActionID: core.ActionID{SpellID: rank.ID},
			Duration: rank.Duration(),
		},
		ShieldStrengthCalculator: func(_ *core.Unit) float64 {
			return rank.EffectN(1).Average(core.CharacterLevel)
		},
		ShouldApplyToResult: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult, _ bool) bool {
			return school == core.SpellSchoolNone || spell.SpellSchool.Matches(school)
		},
	})

	config := spelldata.SpellConfig(&mage.Unit, rank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
	config.ClassSpellMask = mask
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		absorbAura.Activate(sim)
	}
	config.RelatedSelfBuff = absorbAura.Aura

	mage.RegisterSpell(config)
}

func (mage *Mage) registerIceBarrier() {
	if !mage.Talents.IceBarrier {
		return
	}
	mage.registerAbsorbSpell(iceBarrierRank, MageSpellIceBarrier, core.SpellSchoolNone)
}

func (mage *Mage) registerIceBlock() {
	if !mage.Talents.IceBlock {
		return
	}

	iceBlockAura := mage.RegisterAura(spelldata.AuraConfig(iceBlockRank))

	config := spelldata.SpellConfig(&mage.Unit, iceBlockRank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
	config.ClassSpellMask = MageSpellIceBlock
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		// The immunities do not matter to a dps sim; the buff is tracked for the APL.
		iceBlockAura.Activate(sim)
	}
	config.RelatedSelfBuff = iceBlockAura

	mage.RegisterSpell(config)
}

func (mage *Mage) registerCounterspell() {
	config := spelldata.SpellConfig(&mage.Unit, counterspellRank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagNoOnCastComplete
	config.ClassSpellMask = MageSpellCounterspell
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The interrupt itself has no target casting to stop in the sim.
	}

	mage.RegisterSpell(config)
}
