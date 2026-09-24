package mage

var scorchRank = spellData.Scorch.Highest()

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerScorchSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// procChance := []float64{0, 0.33, 0.66, 1}[mage.Talents.ImprovedScorch]
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: scorchRank.ID},
	// 	SpellSchool:    scorchRank.SpellSchool(),
	// 	DefenseType:    scorchRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellScorch,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(scorchRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      scorchRank.GCD(),
	// 			CastTime: scorchRank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplierAdditive: 1,
	// 	BonusCoefficient:         scorchRank.DamageEffect().Coeff(),
	// 	ThreatMultiplier:         1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := scorchRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 		if result.Landed() && mage.Talents.ImprovedScorch > 0 {
	// 			if sim.Proc(procChance, "Improved Scorch") {
	// 				aura := mage.ImprovedScorchAuras.Get(target)
	// 				aura.Activate(sim)
	// 				aura.AddStack(sim)
	// 			}
	// 		}
	// 	},
	//
	// 	RelatedAuraArrays: mage.ImprovedScorchAuras.ToMap(),
	// })
}
