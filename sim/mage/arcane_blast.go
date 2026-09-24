package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneBlastRank = spellData.ArcaneBlast.Highest()

func (mage *Mage) registerArcaneBlastSpell() {

	config := spelldata.SpellConfig(&mage.Unit, arcaneBlastRank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {

	}

	mage.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	//
	// //https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=30451
	// arcaneBlastCoefficient := 0.71399998665
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: arcaneBlastRank.ID},
	// 	SpellSchool:    arcaneBlastRank.SpellSchool(),
	// 	DefenseType:    arcaneBlastRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellArcaneBlast,
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(arcaneBlastRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      arcaneBlastRank.GCD(),
	// 			CastTime: arcaneBlastRank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplier: 1,
	// 	BonusCoefficient: arcaneBlastCoefficient,
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := arcaneBlastRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 		if result.Landed() {
	// 			mage.ArcaneChargesAura.Activate(sim)
	// 			mage.ArcaneChargesAura.AddStack(sim)
	// 		}
	// 	},
	// })
}
