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
	trigger := spelldata.ProcTrigger(&druid.Character, omen, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		druid.ClearcastingAura.Activate(sim)
	}, spelldata.Chance(1))
	// Chance(1) keeps the resolver from building a procs-per-minute manager and leaves core no roll of
	// its own. The chance depends on what hit, so it is drawn here, after the ICD check and before the
	// ICD starts.
	trigger.ExtraCondition = func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
		chance := float64(omen.RPPM) * druid.omenSpeed(spell) / 60
		return sim.Proc(druid.ModifiedProcChance(omen.ClassFlags, chance), trigger.Name)
	}
	druid.MakeProcTriggerAura(trigger)
}

// The speed, in seconds, that turns Omen of Clarity's procs per minute into a chance for this hit.
// An auto attack uses its own swing speed: the paw in Cat and Bear Form, the weapon otherwise. A
// special uses the equipped main-hand weapon in every form. A spell uses its cast time, and an
// instant 1.5 s.
func (druid *Druid) omenSpeed(spell *core.Spell) float64 {
	switch {
	case spell.ProcMask.Matches(core.ProcMaskMeleeWhiteHit):
		return druid.AutoAttacks.MH().SwingSpeed
	case spell.ProcMask.Matches(core.ProcMaskMeleeSpecial):
		return druid.WeaponFromMainHand().SwingSpeed
	case spell.DefaultCast.CastTime > 0:
		return spell.DefaultCast.CastTime.Seconds()
	default:
		return core.GCDDefault.Seconds()
	}
}
