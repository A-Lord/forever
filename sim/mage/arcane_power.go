package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcanePowerRank = spellData.ArcanePower.Highest()

// Arcane Power
// https://www.wowhead.com/forever/spell=12042
//
// "For the next 15 sec, your spells deal 30% more damage while costing 30% more mana to cast."
func (mage *Mage) registerArcanePowerSpell() {
	mage.ArcanePowerAura = mage.RegisterAura(spelldata.AuraConfig(arcanePowerRank))

	// Effect 3 is the SPELLMOD_DOT twin of effect 1's SPELLMOD_DAMAGE, on a mask effect 1 already
	// covers. The damage mod reaches the ticks as well as the hits, so the twin would count twice;
	// the parse folds such a twin only when both masks match, and these do not.
	spelldata.ParseEffects(&mage.Character, mage.ArcanePowerAura, arcanePowerRank, spelldata.SkipEffects(3))

	config := spelldata.SpellConfig(&mage.Unit, arcanePowerRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagNoOnCastComplete))
	config.ClassSpellMask = MageSpellArcanePower

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		mage.ArcanePowerAura.Activate(sim)
	}

	config.RelatedSelfBuff = mage.ArcanePowerAura

	spell := mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: spell,
		Type:  core.CooldownTypeDPS,
	})
}
