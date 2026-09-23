package forever

import "github.com/wowsims/forever/sim/common/shared"

func RegisterPotionsAndConjured() {

	// Potions
	shared.NewSpellDataPotion(118)  // Minor Healing Potion
	shared.NewSpellDataPotion(858)  // Lesser Healing Potion
	shared.NewSpellDataPotion(929)  // Healing Potion
	shared.NewSpellDataPotion(1072) // Full Moonshine
	// unsupported: A_PERIODIC_DAMAGE, which no consumable builder models
	// shared.NewSpellDataPotion(1450) // Potion of Fervor
	shared.NewSpellDataPotion(1710) // Greater Healing Potion
	shared.NewSpellDataPotion(2455) // Minor Mana Potion
	shared.NewSpellDataPotion(2456) // Minor Rejuvenation Potion
	shared.NewSpellDataPotion(3087) // Mug of Shimmer Stout
	shared.NewSpellDataPotion(3384) // Minor Magic Resistance Potion
	shared.NewSpellDataPotion(3385) // Lesser Mana Potion
	shared.NewSpellDataPotion(3827) // Mana Potion
	shared.NewSpellDataPotion(3928) // Superior Healing Potion
	// unsupported: E_DUMMY, which no consumable builder models
	// shared.NewSpellDataPotion(4596) // Minor Discolored Healing Potion
	shared.NewSpellDataPotion(4623) // Lesser Stoneshield Potion
	shared.NewSpellDataPotion(5631) // Rage Potion
	shared.NewSpellDataPotion(5633) // Great Rage Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(6048) // Shadow Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(6049) // Fire Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(6050) // Frost Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(6051) // Holy Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(6052) // Nature Protection Potion
	shared.NewSpellDataPotion(6149) // Greater Mana Potion
	shared.NewSpellDataPotion(9036) // Magic Resistance Potion
	shared.NewSpellDataPotion(9144) // Wildvine Potion
	// unsupported: A_MOD_FLAT_SPELL_DAMAGE_VERSUS, which no consumable builder models; A_MOD_MELEE_ATTACK_POWER_VERSUS, which no consumable builder models
	// shared.NewSpellDataPotion(9224) // Potion of Demonslaying
	// unsupported: A_MOD_STUN, which no consumable builder models
	// shared.NewSpellDataPotion(12190) // Dreamless Sleep Potion
	shared.NewSpellDataPotion(13442) // Mighty Rage Potion
	shared.NewSpellDataPotion(13443) // Superior Mana Potion
	shared.NewSpellDataPotion(13444) // Major Mana Potion
	shared.NewSpellDataPotion(13446) // Major Healing Potion
	shared.NewSpellDataPotion(13455) // Greater Stoneshield Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(13456) // Greater Frost Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(13457) // Greater Fire Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(13458) // Greater Nature Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(13459) // Greater Shadow Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(13460) // Greater Holy Protection Potion
	// unsupported: A_SCHOOL_ABSORB, which no consumable builder models
	// shared.NewSpellDataPotion(13461) // Greater Arcane Protection Potion
	shared.NewSpellDataPotion(17348) // Major Healing Draught
	shared.NewSpellDataPotion(17349) // Superior Healing Draught
	shared.NewSpellDataPotion(17351) // Major Mana Draught
	shared.NewSpellDataPotion(17352) // Superior Mana Draught
	shared.NewSpellDataPotion(18253) // Major Rejuvenation Potion
	shared.NewSpellDataPotion(18839) // Combat Healing Potion
	shared.NewSpellDataPotion(18841) // Combat Mana Potion
	// unsupported: A_MOD_STUN, which no consumable builder models
	// shared.NewSpellDataPotion(20002) // Greater Dreamless Sleep Potion
	shared.NewSpellDataPotion(23578)  // Diet McWeaksauce
	shared.NewSpellDataPotion(23579)  // The McWeaksauce Classic
	shared.NewSpellDataPotion(223913) // Major Healing Potion
	shared.NewSpellDataPotion(223914) // Greater Healing Potion
	// unsupported: E_DUMMY, which no consumable builder models
	// shared.NewSpellDataPotion(241650) // Major Discolored Healing Potion
	// unsupported: E_DUMMY, which no consumable builder models
	// shared.NewSpellDataPotion(247239) // Lesser Discolored Healing Potion
	// unsupported: E_DUMMY, which no consumable builder models
	// shared.NewSpellDataPotion(247240) // Discolored Healing Potion
	// unsupported: E_DUMMY, which no consumable builder models
	// shared.NewSpellDataPotion(247241) // Greater Discolored Healing Potion
	// unsupported: E_DUMMY, which no consumable builder models
	// shared.NewSpellDataPotion(247242) // Superior Discolored Healing Potion
	shared.NewSpellDataPotion(250932) // Minor Spellblasting Potion
	shared.NewSpellDataPotion(250933) // Lesser Spellblasting Potion
	shared.NewSpellDataPotion(250934) // Spellblasting Potion
	shared.NewSpellDataPotion(250935) // Greater Spellblasting Potion
	shared.NewSpellDataPotion(250936) // Superior Spellblasting Potion
	shared.NewSpellDataPotion(250937) // Major Spellblasting Potion
	shared.NewSpellDataPotion(250938) // Minor Frenzy Potion
	shared.NewSpellDataPotion(250939) // Lesser Frenzy Potion
	shared.NewSpellDataPotion(250940) // Frenzy Potion
	shared.NewSpellDataPotion(250941) // Greater Frenzy Potion
	shared.NewSpellDataPotion(250942) // Superior Frenzy Potion
	shared.NewSpellDataPotion(250943) // Major Frenzy Potion
	shared.NewSpellDataPotion(250944) // Minor Mender's Potion
	shared.NewSpellDataPotion(250945) // Lesser Mender's Potion
	shared.NewSpellDataPotion(250946) // Mender's Potion
	shared.NewSpellDataPotion(250947) // Greater Mender's Potion
	shared.NewSpellDataPotion(250948) // Superior Mender's Potion
	shared.NewSpellDataPotion(250949) // Major Mender's Potion
	shared.NewSpellDataPotion(255663) // Windstone
	shared.NewSpellDataPotion(268881) // Perishable Minor Healing Potion
	shared.NewSpellDataPotion(268882) // Perishable Lesser Healing Potion
	shared.NewSpellDataPotion(268883) // Perishable Healing Potion
	shared.NewSpellDataPotion(274935) // Tessa's Tonic
	shared.NewSpellDataPotion(282011) // Restored Healing Potion
	shared.NewSpellDataPotion(282013) // Restored Mana Potion

	// Conjured
	shared.NewSpellDataConjured(1322)  // Fishliver Oil
	shared.NewSpellDataConjured(1703)  // Crystal Basilisk Spine
	shared.NewSpellDataConjured(5509)  // Healthstone
	shared.NewSpellDataConjured(5510)  // Greater Healthstone
	shared.NewSpellDataConjured(5511)  // Lesser Healthstone
	shared.NewSpellDataConjured(5512)  // Minor Healthstone
	shared.NewSpellDataConjured(5513)  // Mana Jade
	shared.NewSpellDataConjured(7676)  // Thistle Tea
	shared.NewSpellDataConjured(8007)  // Mana Citrine
	shared.NewSpellDataConjured(8008)  // Mana Ruby
	shared.NewSpellDataConjured(9421)  // Major Healthstone
	shared.NewSpellDataConjured(11951) // Whipper Root Tuber
	shared.NewSpellDataConjured(11952) // Night Dragon's Breath
	shared.NewSpellDataConjured(12662) // Demonic Rune
	shared.NewSpellDataConjured(14894) // Lily Root
	shared.NewSpellDataConjured(15723) // Tea with Sugar
	shared.NewSpellDataConjured(17747) // Razorlash Root
	shared.NewSpellDataConjured(19004) // Minor Healthstone
	shared.NewSpellDataConjured(19005) // Minor Healthstone
	shared.NewSpellDataConjured(19006) // Lesser Healthstone
	shared.NewSpellDataConjured(19007) // Lesser Healthstone
	shared.NewSpellDataConjured(19008) // Healthstone
	shared.NewSpellDataConjured(19009) // Healthstone
	shared.NewSpellDataConjured(19010) // Greater Healthstone
	shared.NewSpellDataConjured(19011) // Greater Healthstone
	shared.NewSpellDataConjured(19012) // Major Healthstone
	shared.NewSpellDataConjured(19013) // Major Healthstone
	// unsupported: E_TRIGGER_SPELL, which no consumable builder models
	// shared.NewSpellDataConjured(275688) // Power Recycler
}
