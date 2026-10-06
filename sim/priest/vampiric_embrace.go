package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func (priest *Priest) registerVampiricEmbrace(rank *spelldata.Spell) {
	healFraction := rank.EffectN(1).Percent()
	healthMetrics := priest.NewHealthMetrics(core.ActionID{SpellID: rank.ID})

	veAuras := priest.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		return target.RegisterAura(spelldata.AuraConfig(rank, spelldata.Label(rank.Name+" - "+priest.Label)))
	})

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Vampiric Embrace Heal",
		Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
		ClassSpellMask:     PriestShadowSpells, // The tooltip's "Shadow spell damage you deal"; the row's listener names no spells.
		RequireDamageDealt: true,
		CanProcFromProcs:   true,
		ExtraCondition: func(_ *core.Simulation, _ *core.Spell, result *core.SpellResult) bool {
			return veAuras.Get(result.Target).IsActive()
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			priest.GainHealth(sim, result.Damage*healFraction, healthMetrics)
		},
	})

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Magic(core.ProcMaskSpellDamage))

	config.ActionID = core.ActionID{SpellID: rank.ID}
	config.ClassSpellMask = PriestSpellVampiricEmbrace

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, _ *core.Spell) {
		veAuras.Get(target).Activate(sim)
	}
	config.RelatedAuraArrays = veAuras.ToMap()

	priest.VampiricEmbrace = priest.RegisterSpell(config)
}
