package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerInnervateCD() {
	rank := spellData.Innervate.Highest()
	target := druid.GetUnit(druid.SelfBuffs.InnervateTarget)
	if target == nil {
		target = &druid.Unit
	}
	if druid.SelfBuffs.GiftInnervate {
		if dummy := druid.Env.Raid.GetFirstTargetDummy(); dummy != nil {
			target = &dummy.Unit
		}
	}
	aura := buffs.NewSpellDataInnervatesAura(target, true, 0)
	config := spelldata.SpellConfig(&druid.Unit, rank)
	config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !target.HasActiveAuraWithTag(buffs.InnervatesCategory)
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}
	druid.Innervate = druid.RegisterSpell(config)
	druid.AddMajorCooldown(core.MajorCooldown{
		Spell: druid.Innervate.Spell,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(_ *core.Simulation, _ *core.Character) bool {
			return false
		},
	})
}
