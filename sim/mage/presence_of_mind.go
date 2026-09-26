package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var presenceOfMindRank = spellData.PresenceOfMind.Highest()

func (mage *Mage) registerPresenceOfMindSpell() {
	if !mage.Talents.PresenceOfMind {
		return
	}

	var pomSpell *core.Spell

	// The -100% cast time modifier is the row's own, masked to the spells the client names.
	mage.PresenceOfMindAura = mage.RegisterAura(spelldata.AuraConfig(presenceOfMindRank))
	spelldata.ParseEffects(&mage.Character, mage.PresenceOfMindAura, presenceOfMindRank)

	mage.PresenceOfMindAura.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
		pomSpell.CD.Use(sim)
	})
	mage.PresenceOfMindAura.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
		if !spell.Matches(MageSpellsAll ^ (MageSpellInstantCast | MageSpellEvocation)) {
			return
		}
		if spell.DefaultCast.CastTime == 0 {
			return
		}
		aura.Deactivate(sim)
	}

	config := spelldata.SpellConfig(&mage.Unit, presenceOfMindRank)
	config.Flags |= core.SpellFlagNoOnCastComplete | core.SpellFlagAPL
	config.ClassSpellMask = MageSpellPresenceOfMind
	config.Cast.DefaultCast.NonEmpty = true
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		mage.PresenceOfMindAura.Activate(sim)
	}
	config.RelatedSelfBuff = mage.PresenceOfMindAura

	pomSpell = mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: pomSpell,
		Type:  core.CooldownTypeDPS,
	})
}
