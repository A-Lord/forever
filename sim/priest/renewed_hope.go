package priest

import "github.com/wowsims/forever/sim/core"

const priestRenewedHopeSpells = PriestSpellFlashHeal |
	PriestSpellLesserHeal |
	PriestSpellHeal |
	PriestSpellGreaterHeal |
	PriestSpellBindingHeal |
	PriestSpellPenance

func (priest *Priest) withRenewedHope(sim *core.Simulation, target *core.Unit, spell *core.Spell, apply func()) {
	weakenedSoul := priest.renewedHopeTargetAura(target, spell)
	if weakenedSoul == nil {
		apply()
		return
	}

	bonusCrit := spellData.RenewedHope.Rank(priest.Talents.RenewedHope).EffectN(1).Percent()
	// Penance's bolt spells are procs within one cast. They each get the crit bonus, but the cast
	// shortens Weakened Soul only once, when the channel starts.
	if !spell.Flags.Matches(core.SpellFlagProc) {
		priest.reduceWeakenedSoulDuration(sim, weakenedSoul)
	}
	priest.withRenewedHopeCrit(spell, bonusCrit, apply)
}

func (priest *Priest) withRenewedHopeCrit(spell *core.Spell, bonusCrit float64, apply func()) {
	spell.BonusCritPercent += bonusCrit
	defer func() { spell.BonusCritPercent -= bonusCrit }()
	apply()
}

func (priest *Priest) renewedHopeTargetAura(target *core.Unit, spell *core.Spell) *core.Aura {
	if priest.Talents.RenewedHope == 0 || target == nil || target.Type == core.EnemyUnit ||
		spell.ClassSpellMask&priestRenewedHopeSpells == 0 {
		return nil
	}

	aura := priest.weakenedSoulAuras.Get(target)
	if !aura.IsActive() {
		return nil
	}
	return aura
}

func (priest *Priest) reduceWeakenedSoulDuration(sim *core.Simulation, aura *core.Aura) {
	reduction := -spellData.RenewedHope.Rank(priest.Talents.RenewedHope).EffectN(2).TimeValue()
	remaining := aura.RemainingDuration(sim)
	if remaining <= reduction {
		aura.Deactivate(sim)
		return
	}

	baseDuration := aura.Duration
	aura.Duration = remaining - reduction
	aura.Refresh(sim)
	aura.Duration = baseDuration
}
