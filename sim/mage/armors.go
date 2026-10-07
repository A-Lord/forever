package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The Chilled procs on Frost and Ice Armor slow the attacker, which does nothing in the sim. Every
// rank of the three armors is registered, and one armor of any rank is up at a time.
func (mage *Mage) registerArmorSpells() {
	// Frost Warding raises Frost and Ice Armor's armor and resistance.
	warding := 1 + spellData.FrostWarding.EffectAt(1).FractionAt(mage.Talents.FrostWarding)
	// Arcane Shielding's effect 2 raises Mage Armor's resistances (effect 1), so that effect is
	// scaled here; the casting regen (effect 2) is parsed.
	shielding := 1 + spellData.ArcaneShielding.EffectAt(2).FractionAt(mage.Talents.ArcaneShielding)

	spellData.FrostArmor.Each(func(_ int32, rank *spelldata.Spell) {
		mage.registerArmor(rank, MageSpellFrostArmor, stats.Stats{
			stats.Armor: warding * rank.EffectN(1).BaseValue(),
		})
	})

	var iceArmor *core.Aura
	spellData.IceArmor.Each(func(_ int32, rank *spelldata.Spell) {
		aura := mage.registerArmor(rank, MageSpellIceArmor, stats.Stats{
			stats.Armor:           warding * rank.EffectN(1).BaseValue(),
			stats.FrostResistance: warding * rank.EffectN(3).BaseValue(),
		})
		if rank == spellData.IceArmor.Highest() {
			iceArmor = aura
		}
	})

	var mageArmor *core.Aura
	spellData.MageArmor.Each(func(_ int32, rank *spelldata.Spell) {
		resist := shielding * rank.EffectN(1).BaseValue()
		aura := mage.registerArmor(rank, MageSpellMageArmor, stats.Stats{
			stats.ArcaneResistance: resist,
			stats.FireResistance:   resist,
			stats.FrostResistance:  resist,
			stats.NatureResistance: resist,
			stats.ShadowResistance: resist,
		})
		spelldata.ParseEffects(&mage.Character, aura, rank, spelldata.Effects(2))
		if rank == spellData.MageArmor.Highest() {
			mageArmor = aura
		}
	})

	if mage.Options == nil {
		return
	}

	defaultArmor := map[proto.MageArmor]*core.Aura{
		proto.MageArmor_MageArmorFrostArmor: iceArmor,
		proto.MageArmor_MageArmorMageArmor:  mageArmor,
	}[mage.Options.DefaultMageArmor]

	if defaultArmor != nil {
		defaultArmor.BuildPhase = core.CharacterBuildPhaseBuffs
		defaultArmor.ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		})
	}
}

func (mage *Mage) registerArmor(rank *spelldata.Spell, classMask int64, bonus stats.Stats) *core.Aura {
	aura := mage.RegisterAura(spelldata.AuraConfig(rank, spelldata.RankLabel(rank, &mage.Unit))).AttachStatsBuff(bonus)
	aura.NewExclusiveEffect("Mage Armor", true, core.ExclusiveEffect{})

	config := spelldata.SpellConfig(&mage.Unit, rank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
	config.ClassSpellMask = classMask
	config.RelatedSelfBuff = aura

	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !aura.IsActive()
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		aura.Activate(sim)
	}

	mage.RegisterSpell(config)
	return aura
}
