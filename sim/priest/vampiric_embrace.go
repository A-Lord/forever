package priest

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerVampiricEmbrace(rank *spelldata.Spell, cdTimer *core.Timer) {

	// TODO: Forever drops Improved Vampiric Embrace; base heal percent only until we know
	// whether the bonus moved onto another talent.
	healPct := 0.20
	actionId := core.ActionID{SpellID: rank.ID}
	healthMetrics := priest.NewHealthMetrics(actionId)

	veDebuffAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		aura := target.RegisterAura(core.Aura{
			Label:    "Vampiric Embrace",
			ActionID: actionId,
			Duration: time.Second * 30,
		})
		aura.AttachProcTriggerCallback(target, core.ProcTrigger{
			Name:               "Vampiric Embrace Proc",
			Callback:           core.CallbackOnSpellHitTaken | core.CallbackOnPeriodicDamageTaken,
			ClassSpellMask:     PriestShadowSpells,
			RequireDamageDealt: true,
			Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
				priest.GainHealth(sim, result.Damage*healPct, healthMetrics)
			},
		})
		return aura
	})

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = core.ActionID{SpellID: rank.ID}
	config.ClassSpellMask = PriestSpellVampiricEmbrace

	config.Cast.CD = core.Cooldown{
		Timer:    cdTimer,
		Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		veDebuffAuras.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = veDebuffAuras.ToMap()
}
