package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {

	// The Black Book
	core.NewItemEffect(19337, func(agent core.Agent) {
		warlock := agent.(WarlockAgent).GetWarlock()
		duration := time.Second * 30

		petAuras := make([]*core.StatBuffAura, len(warlock.Env.Raid.AllUnits)+1)
		for _, pet := range warlock.Pets {
			petAuras[pet.UnitIndex] = pet.NewTemporaryStatsAura(
				"Blessing of The Black Book",
				core.ActionID{SpellID: 23720},
				stats.Stats{stats.SpellDamage: 200, stats.AttackPower: 325, stats.Armor: 1600},
				duration,
			)
		}

		aura := warlock.RegisterAura(core.Aura{
			Label:    "Blessing of The Black Book",
			ActionID: core.ActionID{ItemID: 19337},
			Duration: duration,
			OnGain: func(aura *core.Aura, sim *core.Simulation) {
				if warlock.ActivePet != nil {
					petAuras[warlock.ActivePet.UnitIndex].Activate(sim)
				}
			},
		})

		spell := warlock.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{ItemID: 19337},
			DefenseType: core.DefenseTypeMagic, // Blessing of the Black Book (23720)
			Flags:       core.SpellFlagNoOnCastComplete,

			Cast: core.CastConfig{
				CD: core.Cooldown{
					Timer:    warlock.NewTimer(),
					Duration: time.Minute * 3,
				},
				SharedCD: core.Cooldown{
					Timer:    warlock.GetOffensiveTrinketCD(),
					Duration: duration,
				},
			},

			ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
				return warlock.ActivePet != nil
			},

			ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
				aura.Activate(sim)
			},
		})

		warlock.AddMajorCooldown(core.MajorCooldown{
			Spell: spell,
			Type:  core.CooldownTypeDPS,
			BuffAura: &core.StatBuffAura{
				Aura:            aura,
				BuffedStatTypes: []stats.Stat{stats.SpellDamage, stats.AttackPower, stats.Armor},
			},
		})
	})

}
