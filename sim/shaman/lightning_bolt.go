package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var LightningBoltRankMap = spellData.LightningBolt

// TODO: To be implemented. Port the TBC Lightning Bolt Spell implementation below; not yet verified against the Forever client.
func (shaman *Shaman) registerLightningBoltSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// shaman.LightningBoltOverloads = make(map[int32]*core.Spell, LightningBoltRankMap.Len())
	// LightningBoltRankMap.Each(func(_ int32, config *spelldata.Spell) {
	// 	shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, false))
	// 	shaman.LightningBoltOverloads[config.RankNumber()] = shaman.RegisterSpell(shaman.newLightningBoltSpellConfig(config, true))
	// })
}

func (shaman *Shaman) newLightningBoltSpellConfig(config *spelldata.Spell, isElementalOverload bool) core.SpellConfig {
	shamConfig := ShamSpellConfig{
		ActionID:            core.ActionID{SpellID: config.ID},
		Rank:                config.RankNumber(),
		IsElementalOverload: isElementalOverload,
		BaseFlatCost:        int32(config.Cost()),
		BonusCoefficient:    config.DamageEffect().Coeff(),
		BaseCastTime:        time.Millisecond * 2500,
	}
	spellConfig := shaman.newElectricSpellConfig(shamConfig)

	spellConfig.ClassSpellMask = core.TernaryInt64(isElementalOverload, SpellMaskLightningBoltOverload, SpellMaskLightningBolt)
	spellConfig.MissileSpeed = 20

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := config.DamageEffect().Average(core.CharacterLevel)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			if !isElementalOverload && result.Landed() && sim.Proc(shaman.GetOverloadChance(), "Lightning Bolt Elemental Overload") {
				shaman.LightningBoltOverloads[config.RankNumber()].Cast(sim, target)
			}

			spell.DealDamage(sim, result)
		})
	}

	return spellConfig
}
