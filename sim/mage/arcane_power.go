package mage

// TODO: To be implemented. TBC body below needs no porting; kept commented until this class's port is reviewed.
func (mage *Mage) registerArcanePowerSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// if !mage.Talents.ArcanePower {
	// 	return
	// }
	//
	// arcanePowerCostMod := mage.AddDynamicMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellsAll,
	// 	FloatValue: .3,
	// 	Kind:       core.SpellMod_PowerCost_Pct_Add,
	// })
	//
	// arcanePowerDmgMod := mage.AddDynamicMod(core.SpellModConfig{
	// 	ClassMask:  MageSpellsAll,
	// 	FloatValue: .3,
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// })
	//
	// var arcanePowerSpell *core.Spell
	// mage.ArcanePowerAura = mage.RegisterAura(core.Aura{
	// 	Label:    "Arcane Power",
	// 	ActionID: core.ActionID{SpellID: 12042},
	// 	Duration: time.Second * 15,
	// })
	//
	// mage.ArcanePowerAura.NewExclusiveEffect("ManaCost", true, core.ExclusiveEffect{
	// 	Priority: 10,
	// 	OnGain: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
	// 		arcanePowerCostMod.Activate()
	// 		arcanePowerDmgMod.Activate()
	// 	},
	// 	OnExpire: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
	// 		arcanePowerCostMod.Deactivate()
	// 		arcanePowerDmgMod.Deactivate()
	// 		arcanePowerSpell.CD.Use(sim)
	// 	},
	// })
	//
	// arcanePowerSpell = mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: 12042},
	// 	DefenseType:    core.DefenseTypeMagic,
	// 	Flags:          core.SpellFlagNoOnCastComplete,
	// 	ClassSpellMask: MageSpellArcanePower,
	// 	ClassFlags:     spellData.ArcanePower.Highest().ClassFlags,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			NonEmpty: true,
	// 		},
	// 		CD: core.Cooldown{
	// 			Timer:    mage.NewTimer(),
	// 			Duration: time.Second * 180,
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		mage.ArcanePowerAura.Activate(sim)
	// 	},
	// 	RelatedSelfBuff: mage.ArcanePowerAura,
	// })
	//
	// mage.AddMajorCooldown(core.MajorCooldown{
	// 	Spell: arcanePowerSpell,
	// 	Type:  core.CooldownTypeDPS,
	// })
}
