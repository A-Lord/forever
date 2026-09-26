package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var arcanePowerRank = spellData.ArcanePower.Highest()

func (mage *Mage) registerArcanePowerSpell() {
	if !mage.Talents.ArcanePower {
		return
	}

	// The +30% damage, +30% cost and +30% dot modifiers are the row's own, masked to the
	// spells the client names.
	mage.ArcanePowerAura = mage.RegisterAura(spelldata.AuraConfig(arcanePowerRank))
	spelldata.ParseEffects(&mage.Character, mage.ArcanePowerAura, arcanePowerRank)

	config := spelldata.SpellConfig(&mage.Unit, arcanePowerRank)
	config.Flags |= core.SpellFlagNoOnCastComplete | core.SpellFlagAPL
	config.ClassSpellMask = MageSpellArcanePower
	config.Cast.DefaultCast.NonEmpty = true
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		mage.ArcanePowerAura.Activate(sim)
	}
	config.RelatedSelfBuff = mage.ArcanePowerAura

	arcanePowerSpell := mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: arcanePowerSpell,
		Type:  core.CooldownTypeDPS,
	})
}
