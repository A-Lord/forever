package mage

var arcaneBlastRank = spellData.ArcaneBlast.Highest()

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcaneBlastSpell() {
	panic("To be implemented")

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
	// 		FlatCost: arcaneBlastRank.Cost(),
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
