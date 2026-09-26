package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var mageArmorRank = spellData.MageArmor.Highest()
var iceArmorRank = spellData.IceArmor.Highest()

// The regen-while-casting share Mage Armor's second effect states (50%).
var mageArmorRegen = mageArmorRank.EffectN(2).Percent()

// Forever carries no Molten Armor: the client's armor spells are Frost Armor, its
// higher-rank Ice Armor, and Mage Armor. The proto option for Molten Armor maps to nothing.
func (mage *Mage) registerArmorSpells() {
	mageArmorEffectCategory := "MageArmors"

	registerArmor := func(rank *spelldata.Spell, mask int64, aura *core.Aura) {
		aura.NewExclusiveEffect(mageArmorEffectCategory, true, core.ExclusiveEffect{})

		config := spelldata.SpellConfig(&mage.Unit, rank)
		config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
		config.ClassSpellMask = mask
		config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
			return !aura.IsActive()
		}
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			aura.Activate(sim)
		}
		config.RelatedSelfBuff = aura
		mage.RegisterSpell(config)
	}

	mageArmor := mage.RegisterAura(spelldata.AuraConfig(mageArmorRank))
	// Effect 2 is A_MOD_MANA_REGEN_INTERRUPT, which the parse table does not carry.
	spelldata.ParseEffects(&mage.Character, mageArmor, mageArmorRank, spelldata.SkipEffects(2))
	mageArmor.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		mage.PseudoStats.SpiritRegenRateCasting += mageArmorRegen
		mage.UpdateManaRegenRates()
	})
	mageArmor.ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
		mage.PseudoStats.SpiritRegenRateCasting -= mageArmorRegen
		mage.UpdateManaRegenRates()
	})
	registerArmor(mageArmorRank, MageSpellMageArmor, mageArmor)

	// Ice Armor's Chilled proc slows melee attackers, which does nothing to a raid boss;
	// only the armor and frost resistance apply.
	iceArmor := mage.RegisterAura(spelldata.AuraConfig(iceArmorRank))
	spelldata.ParseEffects(&mage.Character, iceArmor, iceArmorRank)
	registerArmor(iceArmorRank, MageSpellFrostArmor, iceArmor)

	var defaultArmor *core.Aura
	switch mage.Options.DefaultMageArmor {
	case proto.MageArmor_MageArmorFrostArmor:
		defaultArmor = iceArmor
	case proto.MageArmor_MageArmorMageArmor, proto.MageArmor_MageArmorMoltenArmor:
		// Molten Armor is not in Forever; a config that still names it gets Mage Armor.
		defaultArmor = mageArmor
	}

	if defaultArmor != nil {
		core.MakePermanent(defaultArmor)
		defaultArmor.BuildPhase = core.CharacterBuildPhaseBuffs
	}
}
