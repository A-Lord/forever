package core

// Fills a proc's handler with an extra attack: a copy of the main-hand swing tagged with the
// granting spell, cast count times on the hit's target. The returned OnInit registers the copy, so it
// goes on the proc's aura: the main-hand swing is configured after procs register, and a unit that
// never swings has none to copy, in which case the proc casts nothing. The swing it grants is not a
// hit the proc hears, or one proc would swing forever.
func (unit *Unit) ExtraAttackProc(trigger *ProcTrigger, tagSpellID int32, count int) OnInit {
	var extraAttack *Spell
	condition := trigger.ExtraCondition
	trigger.ExtraCondition = func(sim *Simulation, spell *Spell, result *SpellResult) bool {
		return spell != extraAttack && (condition == nil || condition(sim, spell, result))
	}
	trigger.Handler = func(sim *Simulation, _ *Spell, result *SpellResult) {
		if extraAttack == nil {
			return
		}
		for range count {
			unit.AutoAttacks.MaybeReplaceMHSwing(sim, extraAttack).Cast(sim, result.Target)
		}
	}
	trigger.TriggerImmediately = true

	return func(_ *Aura, _ *Simulation) {
		if unit.AutoAttacks.MHAuto() == nil {
			return
		}
		swing := *unit.AutoAttacks.MHConfig()
		swing.ActionID = swing.ActionID.WithTag(tagSpellID)
		extraAttack = unit.GetOrRegisterSpell(swing)
	}
}
