package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var evocationRank = spellData.Evocation.Highest()

// While channeling, mana regen is increased by 1500% (effect 1) and all of it continues while
// casting (effect 2).
func (mage *Mage) registerEvocation() {
	regenMultiplier := 1 + evocationRank.EffectN(1).Percent()
	castingRegen := evocationRank.EffectN(2).Percent()

	config := spelldata.SpellConfig(&mage.Unit, evocationRank, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = MageSpellEvocation

	config.Hot = spelldata.DotConfig(evocationRank, evocationRank.Effect(dbcenums.A_PERIODIC_DUMMY, 0))
	config.Hot.SelfOnly = true
	config.Hot.OnTick = func(_ *core.Simulation, _ *core.Unit, _ *core.Dot) {}
	config.Hot.Aura.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		mage.PseudoStats.SpiritRegenRateCasting += castingRegen
		mage.MultiplyManaRegenSpeed(sim, regenMultiplier)
	}
	config.Hot.Aura.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		mage.PseudoStats.SpiritRegenRateCasting -= castingRegen
		mage.MultiplyManaRegenSpeed(sim, 1/regenMultiplier)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		spell.SelfHot().Apply(sim)
	}

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: mage.RegisterSpell(config),
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
			return character.CurrentManaPercent() < 0.1
		},
	})
}
