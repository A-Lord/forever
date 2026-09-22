package druid

// Package-level state the commented-out implementations used:
// var tigersFuryRank = spellData.TigersFury.ByID(9846)

// TODO: To be implemented. The ability exists: spells 5217 and 417045 on the Feral Combat line. No rank
// subtext, so no generated table -- pin the id directly.
func (druid *Druid) registerTigersFurySpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// weaponDamageBonus := tigersFuryRank.DamageEffect().Average(core.CharacterLevel)
	//
	// druid.TigersFuryAura = druid.RegisterAura(core.Aura{
	// 	Label:    "Tiger's Fury",
	// 	ActionID: core.ActionID{SpellID: tigersFuryRank.ID},
	// 	Duration: time.Second * 6,
	//
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		druid.AutoAttacks.MH().BaseDamageMin += weaponDamageBonus
	// 		druid.AutoAttacks.MH().BaseDamageMax += weaponDamageBonus
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		druid.AutoAttacks.MH().BaseDamageMin -= weaponDamageBonus
	// 		druid.AutoAttacks.MH().BaseDamageMax -= weaponDamageBonus
	// 	},
	// })
	//
	// druid.TigersFury = druid.RegisterSpell(Cat, core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: tigersFuryRank.ID},
	// 	ClassSpellMask: DruidSpellTigersFury,
	// 	Flags:          core.SpellFlagAPL,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: tigersFuryRank.Cost(),
	// 	},
	// 	Cast: core.CastConfig{
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    druid.NewTimer(),
	// 			Duration: max(tigersFuryRank.Cooldown(), tigersFuryRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
	// 		druid.TigersFuryAura.Activate(sim)
	// 	},
	// })
}
