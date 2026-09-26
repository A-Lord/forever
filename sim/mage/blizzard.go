package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var blizzardRank = spellData.Blizzard.Highest()

// Blizzard's periodic damage is the spell BlizzardTriggered casts each tick.
var blizzardTickRank = spellData.BlizzardTriggered.Highest()
var blizzardTickDamage = blizzardTickRank.DamageEffect()

func (mage *Mage) registerBlizzardSpell() {
	tickLength := blizzardRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()

	tickConfig := spelldata.SpellConfig(&mage.Unit, blizzardTickRank,
		spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc())
	tickConfig.ActionID = core.ActionID{SpellID: blizzardRank.ID}.WithTag(1)
	tickConfig.ClassSpellMask = MageSpellBlizzard
	tickConfig.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.CalcAndDealAoeDamage(sim, blizzardTickDamage.Average(core.CharacterLevel), spell.OutcomeMagicHit)
	}
	blizzardTickSpell := mage.RegisterSpell(tickConfig)

	config := spelldata.SpellConfig(&mage.Unit, blizzardRank, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellBlizzard

	config.Dot = core.DotConfig{
		IsAOE: true,
		Aura: core.Aura{
			Label:    "Blizzard",
			ActionID: core.ActionID{SpellID: blizzardRank.ID},
		},
		NumberOfTicks:        int32(blizzardRank.Duration() / tickLength),
		TickLength:           tickLength,
		AffectedByCastSpeed:  true,
		HasteReducesDuration: true,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			blizzardTickSpell.Cast(sim, target)
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.AOEDot().Apply(sim)
	}

	mage.RegisterSpell(config)
}
