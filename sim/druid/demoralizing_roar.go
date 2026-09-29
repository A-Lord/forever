package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerDemoralizingRoarSpell() {
	rank := spellData.DemoralizingRoar.Highest()
	druid.DemoralizingRoarAuras = druid.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return buffs.NewSpellDataDemoralizingRoarAura(target, true, 0)
	})
	config := spelldata.SpellConfig(&druid.Unit, rank, spelldata.Melee(core.ProcMaskEmpty))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcAndDealOutcome(sim, target, spell.OutcomeMeleeSpecialHit)
		if result.Landed() {
			druid.DemoralizingRoarAuras.Get(target).Activate(sim)
		}
	}
	config.RelatedAuraArrays = druid.DemoralizingRoarAuras.ToMap()
	druid.DemoralizingRoar = druid.RegisterSpell(config)
}

func (druid *Druid) registerDemoralizingRoarAura() {
}
