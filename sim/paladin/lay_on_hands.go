package paladin

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
)

var LayOnHandsRankMap = spellData.LayOnHands

// Lay on Hands
// https://www.wowhead.com/forever/spell=10310
//
// Heals a friendly target for an amount equal to the Paladin's maximum health and restores 550
// of their mana. Drains all of the Paladin's remaining Mana when used, but does not interrupt
// Mana regeneration.
func (paladin *Paladin) registerLayOnHands(row shared.SpellData) {
	manaRestore := shared.SpellDataMin(row.Energize)
	manaMetrics := paladin.NewManaMetrics(core.ActionID{SpellID: row.SpellID})

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: row.SpellID},
		SpellSchool:    row.SpellSchool,
		DefenseType:    row.DefenseType,
		ProcMask:       core.ProcMaskSpellHealing,
		Flags:          core.SpellFlagAPL | core.SpellFlagHelpful,
		ClassSpellMask: SpellMaskLayOnHands,
		Rank:           row.Rank,
		MaxRange:       row.MaxRange,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: row.GCD,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: row.Cooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.Unit.AddMana(sim, -spell.Unit.CurrentMana(), manaMetrics)

			if target.HasManaBar() {
				target.AddMana(sim, manaRestore, manaMetrics)
			}
			spell.CalcAndDealHealing(sim, target, spell.Unit.MaxHealth(), spell.OutcomeHealingCrit)
		},
	})
}
