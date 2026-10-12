package classic

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	// Force Reactive Disk
	core.NewItemEffect(18168, func(agent core.Agent) {
		character := agent.GetCharacter()

		spell := character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 18168},
			SpellSchool: core.SpellSchoolNature,
			DefenseType: core.DefenseTypeMagic, // Force Reactive Disk (22618)
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagPassiveSpell,

			MissileSpeed: float64(spelldata.MustFind(22600).Speed),

			DamageMultiplier: 1,
			ThreatMultiplier: 1,

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				for _, result := range spell.CalcAoeDamage(sim, 25, spell.OutcomeMagicHitAndCrit) {
					spell.DealDamageAfterTravel(sim, result)
				}
			},
		})

		aura := character.MakeProcTriggerAura(core.ProcTrigger{
			Name:     "Force Reactive Disk",
			ProcMask: core.ProcMaskMelee,
			ICD:      time.Second,
			Outcome:  core.OutcomeBlock,
			Callback: core.CallbackOnSpellHitTaken,
			Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
				spell.Cast(sim, result.Target)
			},
		})

		character.ItemSwap.RegisterProc(18168, aura)
	})
}
