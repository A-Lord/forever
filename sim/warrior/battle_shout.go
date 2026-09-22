package warrior

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var battleShoutRank = spellData.BattleShout.Highest()

func (warrior *Warrior) registerBattleShout() {
	baseAttackPower := battleShoutRank.Effect(dbcenums.A_MOD_ATTACK_POWER, 0).Average(core.CharacterLevel)
	attackPower := func() float64 {
		return baseAttackPower + core.TernaryFloat64(warrior.HasBsT2, core.BattleShoutWrathBonus, 0)
	}

	auras := warrior.NewAllyAuraArray(func(unit *core.Unit) *core.Aura {
		aura := core.BattleShoutAura(unit, true, baseAttackPower, battleShoutRank.Duration())
		aura.BuildPhase = core.Ternary(warrior.DefaultShout == proto.WarriorShout_WarriorShoutBattle, core.CharacterBuildPhaseBuffs, core.CharacterBuildPhaseNone)
		return aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
			if ee := aura.ExclusiveEffects[0]; ee.Priority != attackPower() {
				ee.SetPriority(sim, attackPower())
			}
		})
	})
	selfAura := auras.Get(&warrior.Unit)

	config := spelldata.SpellConfig(&warrior.Unit, battleShoutRank, spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ThreatMultiplier = 1
	// TODO: Manual review needed -- spell 25289 carries no threat effect; none is modelled until
	// measured in game.
	config.FlatThreatBonus = 0

	config.ExtraCastCondition = func(sim *core.Simulation, _ *core.Unit) bool {
		return !selfAura.IsActive() || selfAura.ExclusiveEffects[0].Priority <= attackPower()
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
		auras.ActivateAllPlayers(sim)
	}

	config.RelatedAuraArrays = auras.ToMap()

	warrior.BattleShout = warrior.RegisterSpell(config)
}
