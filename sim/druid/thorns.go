package druid

var thornsRank = spellData.Thorns.Highest()

// TODO: To be implemented.
// Self-cast Thorns (rank 7). Reuses the core raid-buff aura, passing the
// druid's own Brambles talent points. If the Thorns raid buff is selected it
// is already registered (buffs apply before Initialize) and wins; otherwise
// the self-cast version uses the druid's actual talent.
func (druid *Druid) registerThornsSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// thornsAura := druid.GetAura("Thorns")
	// if thornsAura == nil {
	// 	// TODO: Forever drops Brambles; the core aura still takes a rank for it, so it is
	// 	// pinned to 0 until we know whether the effect moved onto another talent.
	// 	thornsAura = core.ThornsAura(druid.GetCharacter(), 0)
	// }
	//
	// druid.RegisterSpell(Humanoid, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: thornsRank.ID},
	// 	SpellSchool:    thornsRank.SpellSchool(),
	// 	DefenseType:    thornsRank.DefenseTypeCore(),
	// 	Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
	// 	ClassSpellMask: DruidSpellThorns,
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	MaxRange:       float64(thornsRank.MaxRange),
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: thornsRank.Cost(),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: thornsRank.GCD(),
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		thornsAura.Activate(sim)
	// 	},
	//
	// 	RelatedSelfBuff: thornsAura,
	// })
}
