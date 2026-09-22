package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var berserkerRageRank = spellData.BerserkerRage.Highest()

func (warrior *Warrior) registerBerserkerRage() {
	rageMetrics := warrior.NewRageMetrics(core.ActionID{SpellID: berserkerRageRank.ID})
	rageGain := spellData.ImprovedBerserkerRage.EffectAt(1).TenthsAt(warrior.Talents.ImprovedBerserkerRage)

	// TODO: Ingame test needed -- the client states no amount for the extra rage a hit taken
	// generates; doubled here.
	auraConfig := spelldata.AuraConfig(berserkerRageRank)
	auraConfig.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		warrior.MultiplyDamageTakenRageGen(2)
	}
	auraConfig.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		warrior.MultiplyDamageTakenRageGen(0.5)
	}
	aura := warrior.RegisterAura(auraConfig).AttachFearImmunity()

	config := spelldata.SpellConfig(&warrior.Unit, berserkerRageRank,
		spelldata.Flags(core.SpellFlagAPL|core.SpellFlagCastWhileIncapacitated))

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return warrior.StanceMatches(BerserkerStance)
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		if rageGain > 0 {
			warrior.AddRage(sim, rageGain, rageMetrics)
		}
		aura.Activate(sim)
	}

	config.RelatedSelfBuff = aura

	spell := warrior.RegisterSpell(config)

	warrior.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeSurvival,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			return rageGain > 0 && warrior.CurrentRage()+rageGain <= warrior.MaximumRage()
		},
	})
}
