package warlock

var shadowBoltRank = spellData.ShadowBolt.Highest()
var shadowBoltCoeff = shadowBoltRank.DamageEffect().Coeff()

// TODO: To be implemented. Port the TBC Shadow Bolt implementation below; not yet verified against the Forever client.
func (warlock *Warlock) registerShadowBolt() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	//
	// warlock.RegisterSpell(core.SpellConfig{
	// 	ActionID:       core.ActionID{SpellID: shadowBoltRank.ID},
	// 	SpellSchool:    shadowBoltRank.SpellSchool(),
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	Flags:          core.SpellFlagAPL,
	// 	ClassSpellMask: WarlockSpellShadowBolt,
	// 	MissileSpeed:   float64(shadowBoltRank.Speed),
	//
	// 	ManaCost: core.ManaCostOptions{FlatCost: int32(shadowBoltRank.Cost())},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD:      shadowBoltRank.GCD(),
	// 			CastTime: shadowBoltRank.CastTime(),
	// 		},
	// 	},
	//
	// 	DamageMultiplierAdditive: 1,
	// 	DefenseType:              shadowBoltRank.DefenseTypeCore(),
	// 	ThreatMultiplier:         1,
	// 	BonusCoefficient:         shadowBoltCoeff,
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		dmgRoll := shadowBoltRank.DamageEffect().Average(core.CharacterLevel)
	// 		result := spell.CalcDamage(sim, target, dmgRoll, spell.OutcomeMagicHitAndCrit)
	// 		existingAura := target.GetAurasWithTag("ImprovedShadowBolt")
	//
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 			if len(existingAura) == 0 || existingAura[0].Duration != core.NeverExpires {
	// 				if result.Landed() && result.Outcome.Matches(core.OutcomeCrit) && warlock.Talents.ImprovedShadowBolt > 0 {
	// 					if !warlock.ImpShadowboltAura.IsActive() {
	// 						warlock.ImpShadowboltAura.Activate(sim)
	// 					}
	// 					warlock.ImpShadowboltAura.SetStacks(sim, 4)
	// 				}
	// 			}
	// 		})
	// 	},
	// })
}
