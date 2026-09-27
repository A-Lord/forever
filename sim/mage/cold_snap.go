package mage

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerColdSnapSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// if !mage.Talents.ColdSnap {
	// 	return
	// }
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 11958},
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	Flags:          core.SpellFlagNoOnCastComplete | core.SpellFlagAPL,
	// 	ClassSpellMask: MageSpellColdSnap,
	// 	ClassFlags:     spellData.ColdSnap.Highest().ClassFlags,
	//
	// 	Cast: core.CastConfig{
	// 		CD: core.Cooldown{
	// 			Timer:    mage.NewTimer(),
	// 			Duration: time.Second * 480,
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		if mage.IcyVeins != nil {
	// 			mage.IcyVeins.CD.Reset()
	// 		}
	// 		if mage.SummonWaterElemental != nil {
	// 			mage.SummonWaterElemental.CD.Reset()
	// 		}
	// 	},
	// })
}
