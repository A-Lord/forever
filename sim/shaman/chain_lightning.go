package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var ChainLightningRankMap = spellData.ChainLightning

// TODO: To be implemented. Port the TBC Chain Lightning Spell implementation below; not yet verified against the Forever client.
func (shaman *Shaman) registerChainLightningSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// maxHits := min(3, shaman.Env.TotalTargetCount())
	// sharedCDTimer := shaman.NewTimer()
	// shaman.ChainLightningOverloads = make(map[int32][]*core.Spell, ChainLightningRankMap.Len())
	// ChainLightningRankMap.Each(func(_ int32, config *spelldata.Spell) {
	// 	shaman.newChainLightningSpell(config, false, sharedCDTimer)
	// 	for range maxHits {
	// 		shaman.ChainLightningOverloads[config.RankNumber()] = append(shaman.ChainLightningOverloads[config.RankNumber()], shaman.newChainLightningSpell(config, true, nil))
	// 	}
	// })
	//
}

func (shaman *Shaman) newChainLightningSpell(config *spelldata.Spell, isElementalOverload bool, sharedCDTimer *core.Timer) *core.Spell {
	shamConfig := ShamSpellConfig{
		ActionID:            core.ActionID{SpellID: config.ID},
		Rank:                config.RankNumber(),
		IsElementalOverload: isElementalOverload,
		BaseFlatCost:        int32(config.Cost()),
		BonusCoefficient:    config.DamageEffect().Coeff(),
		SpellSchool:         core.SpellSchoolNature,
		Overloads:           shaman.ChainLightningOverloads,
		BounceReduction:     0.7,
		ClassSpellMask:      core.TernaryInt64(isElementalOverload, SpellMaskChainLightningOverload, SpellMaskChainLightning),
		BaseCastTime:        time.Second * 2,
	}
	spellConfig := shaman.newElectricSpellConfig(shamConfig)
	if !isElementalOverload {
		spellConfig.Cast.CD = core.Cooldown{
			Timer:    sharedCDTimer,
			Duration: time.Second * 6,
		}
	}
	maxHits := int32(3)
	maxHits = min(maxHits, shaman.Env.TotalTargetCount())

	spellConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		curTarget := target

		// Damage calculation and DealDamage are in separate loops so that e.g. a spell power proc
		// can't proc on the first target and apply to the second
		numHits := min(maxHits, shaman.Env.ActiveTargetCount())
		results := make([]*core.SpellResult, numHits)
		for hitIndex := range numHits {
			baseDamage := config.DamageEffect().Average(core.CharacterLevel)
			results[hitIndex] = spell.CalcDamage(sim, curTarget, baseDamage, spell.OutcomeMagicHitAndCrit)

			curTarget = sim.Environment.NextActiveTargetUnit(curTarget)
			spell.DamageMultiplier *= shamConfig.BounceReduction
		}

		for hitIndex := range numHits {
			if !isElementalOverload && results[hitIndex].Landed() && sim.Proc(shaman.GetOverloadChance()/3, "Chain Lightning Elemental Overload") {
				shamConfig.Overloads[config.RankNumber()][hitIndex].Cast(sim, results[hitIndex].Target)
			}
			spell.DealDamage(sim, results[hitIndex])
			spell.DamageMultiplier /= shamConfig.BounceReduction
		}
	}

	return shaman.RegisterSpell(spellConfig)
}
