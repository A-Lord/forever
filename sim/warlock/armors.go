package warlock

// TODO: To be implemented. Port the TBC Armors implementation below; not yet verified against the Forever game files.
func (warlock *Warlock) registerArmors() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// demonArmorBonus := 660.0
	// demonArmorSRBonus := 18.0
	//
	// if warlock.Talents.DemonicAegis > 0 {
	// 	bonusMultiplier := spellData.DemonicAegis.MultiplierAt(warlock.Talents.DemonicAegis)
	//
	// 	demonArmorBonus *= bonusMultiplier
	// 	demonArmorSRBonus *= bonusMultiplier
	// }
	//
	// warlock.DemonArmor = warlock.RegisterAura(core.Aura{
	// 	Label:    "Demon Armor",
	// 	ActionID: core.ActionID{SpellID: 27260},
	// 	Duration: time.Minute * 30,
	// }).AttachStatBuff(stats.Armor, demonArmorBonus).AttachStatBuff(stats.ShadowResistance, demonArmorSRBonus)
	//
	// // Armor selection
	// switch warlock.Options.Armor {
	//
	// case proto.WarlockOptions_DemonArmor:
	// 	core.MakePermanent(warlock.DemonArmor)
	// }
	//
}
