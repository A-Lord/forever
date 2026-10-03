package priest

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

func (priest *Priest) registerHolyTalents() {
	// Tier 1
	priest.applyTwilightFocus()
	priest.applyImprovedRenew()
	priest.applyHolySpecialization()

	// Tier 2
	priest.applySpellWarding()
	priest.applyDivineFury()

	// Tier 3
	priest.applyHolyNova()
	priest.applyBlessedRecovery()
	priest.applyInspiration()

	// Tier 4
	priest.applyHolyReach()
	priest.applyImprovedHealing()
	priest.applySearingLight()
	priest.applyBindingHeal()

	// Tier 5
	priest.applyLitanyOfLight()
	priest.applySpiritOfRedemption()
	priest.applySpiritualGuidance()

	// Tier 6
	priest.applySpiritualHealing()

	// Tier 7
	priest.applyPrayerOfMending()
}

// Gives you a 70% chance to avoid interruption caused by damage while casting any spell.
func (priest *Priest) applyTwilightFocus() {}

// Increases the amount healed by your Renew spell by 5%.
func (priest *Priest) applyImprovedRenew() {}

// Increases the critical effect chance of your Holy spells by 1%.
func (priest *Priest) applyHolySpecialization() {
	if priest.Talents.HolySpecialization == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_BonusCrit_Percent,
		FloatValue: spellData.HolySpecialization.FractionAt(priest.Talents.HolySpecialization),
		ClassMask:  PriestHolySpells,
	})
}

// Reduces all spell damage taken by X%
func (priest *Priest) applySpellWarding() {}

// Reduces the casting time of your Smite, Holy Fire, Heal and Greater Heal spells by x sec
func (priest *Priest) applyDivineFury() {
	if priest.Talents.DivineFury == 0 {
		return
	}

	// -0.1s per rank
	priest.AddStaticMod(core.SpellModConfig{
		Kind:      core.SpellMod_CastTime_Flat,
		TimeValue: time.Duration(spellData.DivineFury.ValueAt(priest.Talents.DivineFury)) * time.Millisecond,
		ClassMask: PriestSpellSmite | PriestSpellHolyFire,
	})
}

// Grants the holy nova spell
func (priest *Priest) applyHolyNova() {
	if !priest.Talents.HolyNova {
		return
	}

	HolyNovaRankMap.Each(func(_ int32, rank *spelldata.Spell) { priest.registerHolyNovaSpell(rank) })
}

var HolyNovaRankMap = spellData.HolyNova

// After being struck by a melee or ranged critical hit, or suffering more than 30% of your maximum Health from a single attack,
//
//	heal 8% of the damage taken over 6 sec. Refreshing this effect carries over any remaining healing.
func (priest *Priest) applyBlessedRecovery() {}

// Your non-periodic critical heals increase your target's Armor by x% for 15 sec.ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (priest *Priest) applyInspiration() {}

// Increases the range of your Smite and Holy Fire spells and the radius of your Prayer of Healing and Holy Nova spells by 10%.
func (priest *Priest) applyHolyReach() {}

// Reduces the Mana cost of your Lesser Heal, Heal, Greater Heal, Penance, and Prayer of Mending spells by x%.
func (priest *Priest) applyImprovedHealing() {
	if priest.Talents.ImprovedHealing == 0 {
		return
	}

	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: spellData.ImprovedHealing.FractionAt(priest.Talents.ImprovedHealing),
		ClassMask:  PriestSpellPenance,
	})

}

// Increases your Holy damage done by (2/5)%, and gives a  (5/10)% chance each time your Holy Fire spell deals periodic damage for your next Holy Nova to cost no Mana.
func (priest *Priest) applySearingLight() {
	if priest.Talents.SearingLight == 0 {
		return
	}

	// +2/5% damage to Smite and Holy Fire
	priest.AddStaticMod(core.SpellModConfig{
		Kind:       core.SpellMod_DamageDone_Flat,
		FloatValue: spellData.SearingLight.Effect(dbcenums.A_MOD_DAMAGE_PERCENT_DONE, 0).FractionAt(priest.Talents.SearingLight),
		ClassMask:  PriestSpellSmite | PriestSpellHolyFire,
	})

	priest.HolyPurposeAura = priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:            "Holy Nova Aura",
		MetricsActionID: core.ActionID{SpellID: 1284536},
		ClassSpellMask:  PriestSpellHolyNova,
		Callback:        core.CallbackOnCastComplete,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell.CurCast.Cost != 0 {
				return
			}
			priest.HolyPurposeAura.Deactivate(sim)
		},
	}).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_PowerCost_Pct,
		FloatValue: -1.0,
		ClassMask:  PriestSpellHolyNova,
	})

	priest.MakeProcTriggerAura(core.ProcTrigger{
		Name:           "Searing Light",
		ClassSpellMask: PriestSpellHolyFire,
		ProcChance:     spellData.SearingLightTriggered.FractionAt(priest.Talents.SearingLight),
		Callback:       core.CallbackOnPeriodicDamageDealt,
		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			priest.HolyPurposeAura.Activate(sim)
		},
	})
}

// Grants binding heal
func (priest *Priest) applyBindingHeal() {}

// When you cast a healing spell, gain Mana equal to 20% of the base cost of the spell if your previous heal was a different spell
func (priest *Priest) applyLitanyOfLight() {}

// Grants spirit of redemption on death
func (priest *Priest) applySpiritOfRedemption() {}

// Increases your spell healing by up to X% of your total Spirit and your spell damage by up to Y% of your total Spirit.
func (priest *Priest) applySpiritualGuidance() {
	if priest.Talents.SpiritualGuidance == 0 {
		return
	}

	//5% of Spirit added to spell damage per rank
	dmgCoeff := spellData.SpiritualGuidance.Effect(dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT, 126).FractionAt(priest.Talents.SpiritualGuidance)
	healCoeff := spellData.SpiritualGuidance.Effect(dbcenums.A_MOD_SPELL_HEALING_OF_STAT_PERCENT, 4).FractionAt(priest.Talents.SpiritualGuidance)
	priest.AddStatDependency(stats.Spirit, stats.SpellDamage, dmgCoeff)
	priest.AddStatDependency(stats.Spirit, stats.HealingPower, healCoeff)
}

// Increases the amount healed by your spells by X%.
func (priest *Priest) applySpiritualHealing() {
	if priest.Talents.SpiritualHealing == 0 {
		return
	}
}

// Grants prayer of Mending
func (priest *Priest) applyPrayerOfMending() {}
