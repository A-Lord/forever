package mage

const blastWaveCoefficient = 0.1930000037

var blastWaveRank = spellData.BlastWave.Highest()

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerBlastWaveSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// if !mage.Talents.BlastWave {
	// 	return
	// }
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: blastWaveRank.ID},
	// 	Flags:          core.SpellFlagAPL,
	// 	SpellSchool:    blastWaveRank.SpellSchool(),
	// 	DefenseType:    blastWaveRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: MageSpellBlastWave,
	//
	// 	BonusCoefficient: blastWaveRank.DamageEffect().Coeff(),
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: blastWaveRank.Cost(),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: blastWaveRank.GCD(),
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    mage.NewTimer(),
	// 			Duration: max(blastWaveRank.Cooldown(), blastWaveRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		baseDamage := blastWaveRank.DamageEffect().Average(core.CharacterLevel)
	// 		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
	// 		//The above returns a result slice if you want to implement the daze on the targets hit
	// 	},
	// })
}
