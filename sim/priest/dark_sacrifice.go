package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// TODO: To be implemented. Dark Sacrifice is content Forever added; the sim has never
// modelled it, so this is a stub rather than a port.
//
// Spell IDs: 1277324, 1277325, 1277326, 1277327, 1277328
// SpellData: spellData.DarkSacrifice (5 ranks)
// SpellMask: PriestSpellDarkSacrifice
//
// Not wired into Initialize() yet -- wiring an unimplemented ability would panic
// on every character of this class. Wire it when it is implemented.
func (priest *Priest) registerDarkSacrifice(rank *spelldata.Spell, darkSacTimer *core.Timer) {
	// panic("To be implemented")

	actionId := core.ActionID{SpellID: rank.ID}
	dummy := rank.Effect(dbcenums.A_PERIODIC_DUMMY, 0)
	healthCost := dummy.Average()
	mana := rank.EnergizeEffect().Average()
	tickLength := dummy.Period()
	numberOfTicks := int(rank.Duration() / tickLength)
	manaMetrics := priest.NewManaMetrics(actionId)

	config := spelldata.SpellConfig(&priest.Unit, rank, spelldata.Flags(core.SpellFlagAPL))

	config.ActionID = actionId
	config.ClassSpellMask = PriestSpellDarkSacrifice

	config.Cast.CD = core.Cooldown{
		Timer:    darkSacTimer,
		Duration: max(rank.Cooldown(), rank.CategoryCooldown()),
	}

	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   rank.Duration(),
			NumTicks: int(numberOfTicks),
			Priority: core.ActionPriorityDOT,
			OnAction: func(s *core.Simulation) {
				character := priest.GetCharacter()
				//Tootltip shows 1640 + Spirit for total
				adjustedMana := float64(mana + character.GetStat(stats.Spirit)/float64(numberOfTicks))

				character.RemoveHealth(sim, healthCost)
				character.AddMana(sim, adjustedMana, manaMetrics)

			},
		})
	}

	darkSac := priest.RegisterSpell(config)
	priest.AddMajorCooldown(core.MajorCooldown{
		Spell: darkSac,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(s *core.Simulation, c *core.Character) bool {
			character := priest.GetCharacter()
			totalRegen := mana * float64(numberOfTicks)
			return character.MaxMana()-character.CurrentMana() >= totalRegen
		},
	})

}
