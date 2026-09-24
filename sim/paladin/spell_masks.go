package paladin

// Spell masks identify the paladin's spells to talents, proc triggers and spell mods.
const (
	SpellMaskNone int64 = 0

	// Abilities
	SpellMaskJudgement int64 = 1 << iota
	SpellMaskHolyStrike
	SpellMaskConsecration
	SpellMaskExorcism
	SpellMaskHammerOfWrath
	SpellMaskHolyWrath
	SpellMaskHolyLight
	SpellMaskFlashOfLight
	SpellMaskLayOnHands
	SpellMaskRighteousFury

	// Talent abilities
	SpellMaskDivineFavor
	SpellMaskHolyShock
	SpellMaskHolyShockHeal
	SpellMaskHolyShield
	SpellMaskHolyShieldProc
	SpellMaskSwiftJudgement
	SpellMaskTemplarsBulwark
	SpellMaskLightsVigil
	SpellMaskLightsVigilStrike

	// Seals
	SpellMaskSealOfRighteousness
	SpellMaskSealOfCommand
	SpellMaskSealOfLight
	SpellMaskSealOfWisdom
	SpellMaskSealOfJustice
	SpellMaskSealOfTheCrusader
	SpellMaskSealOfFury

	// What the seals do on a hit
	SpellMaskSealOfRighteousnessProc
	SpellMaskSealOfCommandProc
	SpellMaskSealOfLightProc
	SpellMaskSealOfWisdomProc
	SpellMaskSealOfFuryProc

	// What Judgement unleashes
	SpellMaskJudgementOfRighteousness
	SpellMaskJudgementOfCommand
	SpellMaskJudgementOfLight
	SpellMaskJudgementOfWisdom
	SpellMaskJudgementOfJustice
	SpellMaskJudgementOfTheCrusader
	SpellMaskJudgementOfFury

	// Auras
	SpellMaskDevotionAura
	SpellMaskRetributionAura
	SpellMaskConcentrationAura
	SpellMaskFireResistanceAura
	SpellMaskFrostResistanceAura
	SpellMaskShadowResistanceAura
)

const (
	SpellMaskAllSeals = SpellMaskSealOfRighteousness |
		SpellMaskSealOfCommand |
		SpellMaskSealOfLight |
		SpellMaskSealOfWisdom |
		SpellMaskSealOfJustice |
		SpellMaskSealOfTheCrusader |
		SpellMaskSealOfFury

	SpellMaskSealProcs = SpellMaskSealOfRighteousnessProc |
		SpellMaskSealOfCommandProc |
		SpellMaskSealOfLightProc |
		SpellMaskSealOfWisdomProc |
		SpellMaskSealOfFuryProc

	SpellMaskAllJudgements = SpellMaskJudgementOfRighteousness |
		SpellMaskJudgementOfCommand |
		SpellMaskJudgementOfLight |
		SpellMaskJudgementOfWisdom |
		SpellMaskJudgementOfJustice |
		SpellMaskJudgementOfTheCrusader |
		SpellMaskJudgementOfFury

	SpellMaskAllAuras = SpellMaskDevotionAura |
		SpellMaskRetributionAura |
		SpellMaskConcentrationAura |
		SpellMaskFireResistanceAura |
		SpellMaskFrostResistanceAura |
		SpellMaskShadowResistanceAura

	// The heals Healing Light, Illumination and Divine Favor name.
	SpellMaskHealingSpells = SpellMaskHolyLight |
		SpellMaskFlashOfLight |
		SpellMaskHolyShockHeal

	// Everything cast without a cast time, for Benediction.
	SpellMaskInstantSpells = SpellMaskAllSeals |
		SpellMaskAllAuras |
		SpellMaskJudgement |
		SpellMaskHolyStrike |
		SpellMaskConsecration |
		SpellMaskExorcism |
		SpellMaskLayOnHands |
		SpellMaskRighteousFury |
		SpellMaskDivineFavor |
		SpellMaskHolyShock |
		SpellMaskHolyShockHeal |
		SpellMaskHolyShield |
		SpellMaskSwiftJudgement |
		SpellMaskTemplarsBulwark
)
