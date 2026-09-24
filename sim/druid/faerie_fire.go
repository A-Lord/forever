package druid

// Package-level state the commented-out implementations used:
// var faerieFireFeralRank = spellData.FaerieFireFeral.ByID(27011)

var faerieFireRank = spellData.FaerieFire.Highest()

// TODO: To be implemented.
func (druid *Druid) registerFaerieFireSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// auras := druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	// The druid's own copy of the aura; the client has no Improved Faerie
	// 	// Fire node, so there are no talent points to pass.
	// 	return buffs.FaerieFireAura(target, true, 0)
	// })
	//
	// druid.FaerieFire = druid.RegisterSpell(core.SpellConfig{
	// 	ClassSpellMask:  DruidSpellFaerieFire,
	// 	ActionID:        core.ActionID{SpellID: faerieFireRank.ID},
	// 	CastRequirement: faerieFireRank.CastRequirement(),
	// 	SpellSchool:     faerieFireRank.SpellSchool(),
	// 	DefenseType:     faerieFireRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskSpellDamage,
	// 	Flags:           core.SpellFlagAPL,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(faerieFireRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: faerieFireRank.GCD(),
	// 		},
	// 	},
	//
	// 	ThreatMultiplier: 1,
	// 	FlatThreatBonus:  132,
	// 	MaxRange:         float64(faerieFireRank.MaxRange),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	//
	// 		if result.Landed() {
	// 			auras.Get(target).Activate(sim)
	// 		}
	// 	},
	//
	// 	RelatedAuraArrays: auras.ToMap(),
	// })
}

// TODO: uncalled -- Forever drops the Faerie Fire (Feral) talent; re-gate before wiring
// back into RegisterFeralCatSpells/RegisterFeralTankSpells.
// TODO: To be implemented. Forever has no separate Feral version. Plain Faerie Fire exists on the
// Balance line as spells 770, 778, 9749, 9907, and the feral variant folded into it.
func (druid *Druid) registerFaerieFireFeralSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// druid.FaerieFireAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
	// 	// The druid's own copy of the aura; the client has no Improved Faerie
	// 	// Fire node, so there are no talent points to pass.
	// 	return buffs.FaerieFireAura(target, true, 0)
	// })
	//
	// druid.FaerieFireFeral = druid.RegisterSpell(core.SpellConfig{
	// 	ClassSpellMask:  DruidSpellFaerieFireFeral,
	// 	ActionID:        core.ActionID{SpellID: faerieFireFeralRank.ID},
	// 	CastRequirement: faerieFireFeralRank.CastRequirement(),
	// 	SpellSchool:     faerieFireFeralRank.SpellSchool(),
	// 	DefenseType:     faerieFireFeralRank.DefenseTypeCore(),
	// 	ProcMask:        core.ProcMaskSpellDamage,
	// 	Flags:           core.SpellFlagAPL,
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: faerieFireFeralRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 		CD: core.Cooldown{
	// 			Timer:    druid.NewTimer(),
	// 			Duration: max(faerieFireFeralRank.Cooldown(), faerieFireFeralRank.CategoryCooldown()),
	// 		},
	// 	},
	//
	// 	ThreatMultiplier: 1,
	// 	FlatThreatBonus:  132,
	// 	MaxRange:         float64(faerieFireFeralRank.MaxRange),
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	//
	// 		if result.Landed() {
	// 			druid.FaerieFireAuras.Get(target).Activate(sim)
	// 		}
	// 	},
	//
	// 	RelatedAuraArrays: druid.FaerieFireAuras.ToMap(),
	// })
}
