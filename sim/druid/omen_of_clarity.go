package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (druid *Druid) registerOmenOfClarity() {
	clearRow := spellData.OmenOfClarityTriggered.Highest()
	druid.ClearcastingAura = druid.RegisterAura(spelldata.AuraConfig(clearRow))
	spelldata.ParseEffects(&druid.Character, druid.ClearcastingAura, clearRow)
	mask := clearRow.EffectN(1).ClassFlags
	druid.ClearcastingAura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if mask.Matches(spell.ClassFlags) {
			aura.Deactivate(sim)
		}
	}

	omen := spellData.OmenOfClarity.Highest()
	druid.MakeProcTriggerAura(spelldata.ProcTrigger(&druid.Character, omen, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		druid.ClearcastingAura.Activate(sim)
	}))
}
