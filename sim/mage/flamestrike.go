package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Forever's Flamestrike tops out at rank 6, so only that remains.
var FlameStrikeRankMap = spelldata.Ranked(spellData.Flamestrike.Rank(6).ID)

// Flamestrike
// https://www.wowhead.com/forever/spell=10216
//
// Calls down a pillar of fire, burning all enemies within the area for X Fire damage and an additional
// Y Fire damage over 8 sec. Only one Flamestrike can be active per Mage at a time.
func (mage *Mage) registerFlamestrike(rankConfig *spelldata.Spell) {
	// Flamestrike's periodic damage is the spell FlamestrikeTriggered casts each tick, at the
	// same rank: each rank of Flamestrike names a different rank of that family.
	tick := spellData.FlamestrikeTriggered.Rank(rankConfig.RankNumber()).DamageEffect()
	tickLength := rankConfig.Effect(dbcenums.A_PERIODIC_DUMMY, 0).Period()

	config := spelldata.SpellConfig(&mage.Unit, rankConfig, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = MageSpellFlamestrike

	config.Dot = core.DotConfig{
		IsAOE: true,
		Aura: core.Aura{
			ActionID: core.ActionID{SpellID: rankConfig.ID},
			Label:    "Flamestrike" + mage.Label + " " + rankConfig.Rank,
		},
		NumberOfTicks:    int32(rankConfig.Duration() / tickLength),
		TickLength:       tickLength,
		BonusCoefficient: tick.Coeff(),
		OnTick: func(sim *core.Simulation, _ *core.Unit, dot *core.Dot) {
			dot.Spell.CalcAndDealPeriodicAoeDamage(sim, tick.Average(core.CharacterLevel), dot.OutcomeTickMagicHit)
		},
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		baseDamage := rankConfig.DamageEffect().Roll(sim, core.CharacterLevel)
		spell.CalcAndDealAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
		spell.AOEDot().Apply(sim)
	}

	spell := mage.RegisterSpell(config)

	mage.Flamestrike = append(mage.Flamestrike, spell)
}
