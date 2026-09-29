package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerBashSpell() {
	druid.Bash = druid.registerBearUtility(spellData.Bash.Highest(), true)
}

func (druid *Druid) registerGrowlSpell() {
	druid.Growl = druid.registerBearUtility(spellData.Growl.Highest(), false)
}

func (druid *Druid) registerChallengingRoarSpell() {
	druid.ChallengingRoar = druid.registerBearUtility(spellData.ChallengingRoar.Highest(), false)
}

// A bear spell with no damage contribution. The aura is the debuff the cast leaves, so an APL can call it.
func (druid *Druid) registerBearUtility(rank *spelldata.Spell, onTarget bool) *DruidSpell {
	var auras core.AuraArray
	if onTarget {
		auras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
			return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label(rank.Name)))
		})
	}
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		if auras != nil {
			auras.Get(target).Activate(sim)
		}
	}
	if auras != nil {
		config.RelatedAuraArrays = auras.ToMap()
	}
	return druid.RegisterSpell(config)
}
