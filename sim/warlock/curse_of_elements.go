package warlock

var curseOfElementsRank = spellData.CurseOfTheElements.Highest()

// TODO: To be implemented. The body below builds the spell around the generated
// aura and has not been checked against the client past the aura itself.
func (warlock *Warlock) registerCurseOfElements() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// warlock.CurseOfElementsAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	// The warlock's own copy of the aura. Malediction's modifier masks do
	// 	// not cover this curse, so there are no talent points to pass.
	// 	return buffs.CurseOfElementsAura(target, true, 0)
	// })
	// warlock.CurseOfElements = warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: curseOfElementsRank.ID},
	// 	SpellSchool:    curseOfElementsRank.SpellSchool(),
	// 	DefenseType:    curseOfElementsRank.DefenseTypeCore(),
	// 	ProcMask:       core.ProcMaskEmpty,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellCurseOfElements,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(curseOfElementsRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: curseOfElementsRank.GCD(),
	// 		},
	// 	},
	//
	// 	ThreatMultiplier: 1,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcOutcome(sim, target, spell.OutcomeMagicHit)
	// 		if result.Landed() {
	// 			warlock.DeactivateOtherCurses(sim, spell, target)
	// 			warlock.CurseOfElementsAuras.Get(target).Activate(sim)
	// 		}
	//
	// 		spell.DealOutcome(sim, result)
	// 	},
	//
	// 	RelatedAuraArrays: warlock.CurseOfElementsAuras.ToMap(),
	// })
}
