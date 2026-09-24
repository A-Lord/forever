package druid

// TODO: To be implemented.
func (druid *Druid) registerBarkskin() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// actionId := core.ActionID{SpellID: 22812}
	//
	// barkskinAura := druid.RegisterAura(core.Aura{
	// 	Label:    "Barkskin",
	// 	ActionID: actionId,
	// 	Duration: time.Second * 12,
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		druid.PseudoStats.DamageTakenMultiplier *= 0.8
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		druid.PseudoStats.DamageTakenMultiplier /= 0.8
	// 	},
	//
	// 	// pushback?
	// })
	//
	// druid.Barkskin = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        actionId,
	// 	CastRequirement: spellData.Barkskin.Highest().CastRequirement(),
	// 	DefenseType:     core.DefenseTypeMagic,
	// 	Flags:           core.SpellFlagAPL,
	//
	// 	Cast: core.CastConfig{
	// 		CD: core.Cooldown{
	// 			Timer:    druid.NewTimer(),
	// 			Duration: time.Second * 60,
	// 		},
	// 		DefaultCast: core.Cast{
	// 			GCD: core.GCDDefault,
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		barkskinAura.Activate(sim)
	//
	// 		if sim.CurrentTime > 0 {
	// 			druid.AutoAttacks.StopMeleeUntil(sim, sim.CurrentTime)
	// 		}
	// 	},
	//
	// 	RelatedSelfBuff: barkskinAura,
	// })
	//
	// druid.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: druid.Barkskin.Spell,
	// 	Type:  core.CooldownTypeSurvival,
	// 	ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
	// 		return false // Require manual usage
	// 	},
	// })
}
