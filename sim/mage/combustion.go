package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var combustionRank = spellData.Combustion.Highest()
var combustionBuffRank = spellData.CombustionTriggered.Highest()

// Combustion
// https://www.wowhead.com/forever/spell=11129
//
// "When activated, this spell causes each of your Fire damage spell hits to increase your critical
// strike chance with Fire damage spells by 10%. This effect lasts until you have caused 4
// non-periodic critical strikes with Fire spells."
//
// Two auras: 11129 counts the crits left in its stacks (its ProcCharges), and 28682 carries the
// crit, a SPELLMOD_CRITICAL_CHANCE per stack on its own class mask, up to 10 stacks.
func (mage *Mage) registerCombustionSpell() {
	if !mage.Talents.Combustion {
		return
	}

	buffAura := mage.RegisterAura(spelldata.AuraConfig(combustionBuffRank, spelldata.Label("Combustion Crit")))
	spelldata.ParseEffects(&mage.Character, buffAura, combustionBuffRank)

	var combustionSpell *core.Spell

	chargesConfig := spelldata.AuraConfig(combustionRank, spelldata.Label("Combustion"))

	chargesConfig.OnGain = func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, aura.MaxStacks)
		buffAura.Activate(sim)
	}

	// The row is Disabled While Active (Attr0 0x2000000): the cooldown starts when the effect ends.
	chargesConfig.OnExpire = func(aura *core.Aura, sim *core.Simulation) {
		buffAura.Deactivate(sim)
		combustionSpell.CD.Use(sim)
		mage.UpdateMajorCooldowns()
	}

	// Each landed Fire hit adds a stack, a crit included; a crit then spends one of the charges.
	// Periodic damage never reaches OnSpellHitDealt, so DoT ticks and Ignite neither stack nor spend.
	chargesConfig.OnSpellHitDealt = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		if !spell.SpellSchool.Matches(core.SpellSchoolFire) || !result.Landed() {
			return
		}

		buffAura.AddStack(sim)

		if result.DidCrit() {
			aura.RemoveStack(sim)
			if aura.GetStacks() == 0 {
				aura.Deactivate(sim)
			}
		}
	}

	chargesAura := mage.RegisterAura(chargesConfig)

	// An off-GCD cast of its own, so it does not consume Arcane Charges.
	config := spelldata.SpellConfig(&mage.Unit, combustionRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagNoOnCastComplete))
	config.ClassSpellMask = MageSpellCombustion

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return !chargesAura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		chargesAura.Activate(sim)
	}

	config.RelatedSelfBuff = chargesAura

	combustionSpell = mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: combustionSpell,
		Type:  core.CooldownTypeDPS,
	})
}
