package rogue

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var sliceAndDiceRank = spellData.SliceAndDice.Highest()

func (rogue *Rogue) registerSliceAndDice() {
	config := spelldata.SpellConfig(&rogue.Unit, sliceAndDiceRank, spelldata.Melee(core.ProcMaskEmpty))
	config.ExtraCastCondition = rogue.FinisherExtraCastCondition
	talentModifier := spellData.ImprovedSliceAndDice.MultiplierAt(rogue.Talents.ImprovedSliceAndDice)

	sndAura := rogue.RegisterAura(spelldata.AuraConfig(sliceAndDiceRank))
	spelldata.ParseEffects(&rogue.Character, sndAura, sliceAndDiceRank)

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		rogue.BreakStealth(sim) // This was true in Vanilla, but needs to be tested as future expansions may change this behavior.
		baseDuration := sliceAndDiceRank.Duration().Seconds()
		pointDuration := 3.0 * float64(rogue.ComboPoints())
		fullDuration := (baseDuration + pointDuration) * talentModifier
		sndAura.Duration = time.Duration(fullDuration*1000) * time.Millisecond
		rogue.ApplyFinisher(sim, spell)
		sndAura.Activate(sim)
	}

	rogue.SliceAndDice = rogue.RegisterSpell(config)

	// The TBC implementation, kept for the port:
	// actionID := core.ActionID{SpellID: sliceAndDiceRank.ID}
	//
	// rogue.SliceAndDiceBonusFlat = 0.3
	// rogue.sliceAndDiceDurations = [6]time.Duration{
	// 	0,
	// 	time.Duration(time.Second * 9),
	// 	time.Duration(time.Second * 12),
	// 	time.Duration(time.Second * 15),
	// 	time.Duration(time.Second * 18),
	// 	time.Duration(time.Second * 21),
	// }
	//
	// var sliceAndDiceMod float64
	// rogue.SliceAndDiceAura = rogue.RegisterAura(core.Aura{
	// 	Label:    "Slice and Dice",
	// 	ActionID: actionID,
	// 	// This will be overridden on cast, but set a non-zero default so it doesn't crash when used in APL prepull
	// 	Duration: rogue.sliceAndDiceDurations[5],
	// 	OnGain: func(aura *core.Aura, sim *core.Simulation) {
	// 		sliceAndDiceMod = 1 + rogue.SliceAndDiceBonusFlat
	// 		rogue.MultiplyMeleeSpeed(sim, sliceAndDiceMod)
	// 		if sim.Log != nil {
	// 			rogue.Log(sim, "[DEBUG]: Slice and Dice attack speed mod: %v", sliceAndDiceMod)
	// 		}
	// 	},
	// 	OnExpire: func(aura *core.Aura, sim *core.Simulation) {
	// 		rogue.MultiplyMeleeSpeed(sim, 1/sliceAndDiceMod)
	// 	},
	// })
	//
	// rogue.SliceAndDice = rogue.RegisterSpell(core.SpellConfig{
	// 	ActionID:       actionID,
	// 	Flags:          SpellFlagFinisher | core.SpellFlagAPL,
	// 	MetricSplits:   6,
	// 	ClassSpellMask: RogueSpellSliceAndDice,
	// 	ClassFlags:     sliceAndDiceRank.ClassFlags,
	//
	// 	EnergyCost: core.EnergyCostOptions{
	// 		Cost: int32(sliceAndDiceRank.Cost()),
	// 	},
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: sliceAndDiceRank.GCD(),
	// 		},
	// 		IgnoreHaste: true,
	// 		ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
	// 			spell.SetMetricsSplit(rogue.ComboPoints())
	// 		},
	// 	},
	// 	ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
	// 		return rogue.ComboPoints() > 0
	// 	},
	//
	// 	ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
	// 		comboPoints := rogue.ComboPoints()
	// 		rogue.ApplyFinisher(sim, spell)
	// 		spell.RelatedSelfBuff.Deactivate(sim)
	// 		spell.RelatedSelfBuff.Duration = rogue.getSliceDuration(comboPoints)
	// 		spell.RelatedSelfBuff.Activate(sim)
	// 	},
	//
	// 	RelatedSelfBuff: rogue.SliceAndDiceAura,
	// })
}
