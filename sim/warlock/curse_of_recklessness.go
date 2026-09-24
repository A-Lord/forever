package warlock

var curseOfRecklessnessRank = spellData.CurseOfRecklessness.Highest()

// TODO: To be implemented. The body below builds the spell around the generated
// aura and has not been checked against the client past the aura itself.
func (warlock *Warlock) registerCurseOfRecklessness() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// warlock.CurseOfRecklessnessAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	return buffs.CurseOfRecklessnessAura(target, true, 0)
	// })
	// warlock.CurseOfRecklessness = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: curseOfRecklessnessRank.ID},
	// 	SpellSchool:    curseOfRecklessnessRank.SpellSchool(),
	// 	DefenseType:    curseOfRecklessnessRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellCurseOfRecklessness,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(curseOfRecklessnessRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: curseOfRecklessnessRank.GCD(),
	// 		},
	// 	},
	//
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	// 		if result.Landed() {
	// 			warlock.DeactivateOtherCurses(sim, spell, target)
	// 			warlock.CurseOfRecklessnessAuras.Get(target).Activate(sim)
	// 		}
	//
	// 		spell.DealOutcome(sim, result)
	// 	},
	//
	// 	RelatedAuraArrays: warlock.CurseOfRecklessnessAuras.ToMap(),
	// })
}
