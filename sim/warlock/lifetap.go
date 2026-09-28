package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Converts the row's amount + Spirit of health into as much mana, 1:1, with Improved Life Tap raising
// the mana alone.
// Improved Life Tap's row is a dummy only this formula reads, so the talent is read here rather than parsed.
func (warlock *Warlock) registerLifeTap() {
	rankedSpell := spellData.LifeTap.ByID(11689)
	baseAmount := rankedSpell.EffectN(1).Average(warlock.Level)
	improvedLifeTapMult := 1 + spellData.ImprovedLifeTap.FractionAt(warlock.Talents.ImprovedLifeTap)

	manaMetrics := warlock.NewManaMetrics(core.ActionID{SpellID: rankedSpell.ID})

	// Demonic Energies: the demon out gains a share of the mana
	var petManaMetrics *core.ResourceMetrics
	if warlock.lifeTapPetShare > 0 && warlock.ActivePet != nil {
		petManaMetrics = warlock.ActivePet.NewManaMetrics(core.ActionID{SpellID: rankedSpell.ID})
	}

	config := spelldata.SpellConfig(&warlock.Unit, rankedSpell, spelldata.Flags(core.SpellFlagAPL))
	config.ClassSpellMask = WarlockSpellLifeTap

	// Pay the health and gain the mana, both base + Spirit read at the cast
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		amount := baseAmount + warlock.GetStat(stats.Spirit)
		warlock.RemoveHealth(sim, amount)
		warlock.AddMana(sim, amount*improvedLifeTapMult, manaMetrics)

		if petManaMetrics != nil && warlock.ActivePet.IsEnabled() {
			warlock.ActivePet.AddMana(sim, amount*improvedLifeTapMult*warlock.lifeTapPetShare, petManaMetrics)
		}
	}

	warlock.LifeTap = warlock.RegisterSpell(config)
}
