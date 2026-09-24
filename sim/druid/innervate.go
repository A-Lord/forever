package druid

// TODO: To be implemented. The spirit regen the raid config's innervate applies
// lives in the driver next to the generated aura and has no exported entry
// point, so the druid's own cast needs one before this body comes back.
func (druid *Druid) registerInnervateCD() {
	panic("To be implemented")

	// The body the port needs:
	// innervateTarget := druid.GetUnit(druid.SelfBuffs.InnervateTarget)
	// if innervateTarget == nil {
	// 	innervateTarget = &druid.Unit
	// }
	// innervateTargetChar := druid.Env.Raid.GetPlayerFromUnit(innervateTarget).GetCharacter()
	//
	// actionID := core.ActionID{SpellID: 29166, Tag: druid.Index}
	// var innervateSpell *DruidSpell
	//
	// innervateCD := buffs.InnervatesCooldown()
	//
	// amount := 0.05
	// if innervateTarget == &druid.Unit {
	// 	// TODO: Forever drops Dreamstate; untalented (self-cast base amount) until we
	// 	// know whether the effect moved onto another talent.
	// 	amount = 0.2
	// }
	//
	// var innervateAura = buffs.InnervatesAura(innervateTarget, true, 0)
	//
	// innervateSpell = druid.RegisterSpell(core.SpellConfig{
	// 	ActionID:        actionID,
	// 	CastRequirement: spellData.Innervate.Highest().CastRequirement(),
	// 	DefenseType:     core.DefenseTypeMagic,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: core.GCDDefault,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    druid.NewTimer(),
	// 			Duration: innervateCD,
	// 		},
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		// If target already has another innervate, don't cast.
	// 		return !innervateTarget.HasActiveAuraWithTag(buffs.InnervatesCategory)
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		innervateAura.Activate(sim)
	// 	},
	// })
	//
	// druid.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: innervateSpell.Spell,
	// 	Type:  core.CooldownTypeMana,
	// 	ShouldActivate: func(sim *core.Simulation, character *core.Character) bool {
	// 		// Require manual APL usage
	// 		return false
	// 	},
	// })
}
