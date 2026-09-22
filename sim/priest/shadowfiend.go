package priest

// Package-level state the commented-out implementations used:
// var shadowfiendRank = spellData.Shadowfiend.ByID(34433)

// TODO: To be implemented. The ability exists: spell 401977 on the Shadow Magic line. No rank subtext,
// so no generated table -- pin the id directly. shadowfiend_pet.go already implements the pet.
func (priest *Priest) registerShadowfiendSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// actionID := core.ActionID{SpellID: shadowfiendRank.ID}
	//
	// // Timeline aura
	// priest.ShadowfiendAura = priest.RegisterAura(core.Aura{
	// 	ActionID: actionID,
	// 	Label:    "Shadowfiend",
	// 	Duration: time.Second * 15,
	// })
	//
	// priest.Shadowfiend = priest.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID,
	// 	SpellSchool:    core.SpellSchoolShadow,
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: PriestSpellShadowFiend,
	// 	Rank:           shadowfiendRank.RankNumber(),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		BaseCostPercent: 6,
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: shadowfiendRank.GCD(),
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    priest.NewTimer(),
	// 			Duration: time.Minute * 5,
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		priest.ShadowfiendPet.EnableWithTimeout(sim, priest.ShadowfiendPet, spell.RelatedSelfBuff.Duration)
	// 		spell.RelatedSelfBuff.Activate(sim)
	// 	},
	//
	// 	RelatedSelfBuff: priest.ShadowfiendAura,
	// })
}
