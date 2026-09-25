package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcaneMissilesRank = spellData.ArcaneMissiles.Highest()
var triggered = spellData.ArcaneMissilesTriggered.Highest()
var arcaneMissilesBaseDamage = triggered.DamageEffect().Average(core.CharacterLevel)

func (mage *Mage) registerArcaneMissilesSpell() {
	tickConfig := spelldata.SpellConfig(&mage.Unit, triggered, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.RemoveFlags(core.SpellFlagAPL))
	tickConfig.ClassSpellMask = MageSpellArcaneMissilesTick
	tickConfig.MissileSpeed = 20
	tickConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealDamage(sim, target, arcaneMissilesBaseDamage, spell.OutcomeTickMagicHitAndCrit)
		spell.SpellMetrics[result.Target.UnitIndex].Casts--
	}

	tickSpell := mage.RegisterSpell(tickConfig)

	spellConfig := spelldata.SpellConfig(&mage.Unit, arcaneMissilesRank, spelldata.Magic(core.ProcMaskSpellDamage))
	spellConfig.ClassSpellMask = MageSpellArcaneMissilesCast

	dotConfig := spelldata.DotConfig(arcaneMissilesRank, arcaneMissilesRank.PeriodicEffect())
	dotConfig.OnTick = func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
		tickSpell.Cast(sim, target)
	}

	spellConfig.Dot = dotConfig

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)

		tickSpell.SpellMetrics[target.UnitIndex].Casts++
		if result.Landed() {
			spell.Dot(target).Apply(sim)
			tickSpell.SpellMetrics[target.UnitIndex].Hits++
		}
	}

	mage.RegisterSpell(spellConfig)

	//Below i'm keeping for visibility and confirmation while using spelldata methods

	// The TBC implementation, kept for the port:
	// // Values found at https://wago.tools/db2/SpellEffect?build=2.5.5.65295&filter%5BSpellID%5D=exact%253A7268
	// arcaneMissilesCoefficient := 0.28600001335
	// actionID := core.ActionID{SpellID: arcaneMissilesRank.ID}
	//
	// arcaneMissilesTickSpell := mage.GetOrRegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID.WithTag(1),
	// 	SpellSchool:    core.SpellSchoolArcane,
	// 	DefenseType:    core.DefenseTypeMagic, // Arcane Missile (7268), the real damage spell behind this tick's ActionID (38699)
	// 	ProcMask:       core.ProcMaskSpellDamage,
	// 	ClassSpellMask: MageSpellArcaneMissilesTick,
	// 	// The channel itself carries no speed - the missile spell does, and it is not a ranked row.
	// 	MissileSpeed: 20,
	//
	// 	DamageMultiplier: 1,
	// 	ThreatMultiplier: 1,
	// 	BonusCoefficient: arcaneMissilesCoefficient,
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcDamage(sim, target, 265, spell.OutcomeTickMagicHitAndCrit)
	// 		spell.SpellMetrics[result.Target.UnitIndex].Casts--
	// 		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
	// 			spell.DealDamage(sim, result)
	// 		})
	// 	},
	// })
	//
	// mage.RegisterSpell(core.SpellConfig{
	// 	ActionID:         actionID, // Real SpellID: 5143
	// 	SpellSchool:      core.SpellSchoolArcane,
	// 	ProcMask:         core.ProcMaskSpellDamage,
	// 	Flags:            core.SpellFlagChanneled | core.SpellFlagAPL,
	// 	ClassSpellMask:   MageSpellArcaneMissilesCast,
	// 	DamageMultiplier: 0,
	//
	// 	ManaCost: core.ManaCostOptions{
	// 		FlatCost: int32(arcaneMissilesRank.Cost()),
	// 	},
	//
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: arcaneMissilesRank.GCD(),
	// 		},
	// 	},
	//
	// 	Dot: core.DotConfig{
	// 		Aura: core.Aura{
	// 			Label: "ArcaneMissiles",
	// 		},
	// 		NumberOfTicks:        5,
	// 		TickLength:           time.Second,
	// 		HasteReducesDuration: true,
	// 		AffectedByCastSpeed:  true,
	// 		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
	// 			arcaneMissilesTickSpell.Cast(sim, target)
	// 		},
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMagicHit)
	// 		//Casts is out here and not in Landed() to refer to the number of attempted casts
	// 		arcaneMissilesTickSpell.SpellMetrics[target.UnitIndex].Casts++
	// 		if result.Landed() {
	// 			spell.Dot(target).Apply(sim)
	// 			arcaneMissilesTickSpell.SpellMetrics[target.UnitIndex].Hits++
	// 		}
	// 	},
	// })
}
