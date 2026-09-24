package paladin

import (
	"time"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{18, 16, 18}

type Paladin struct {
	core.Character

	Talents *proto.PaladinTalents

	Forbearance       *core.Aura
	RighteousFuryAura *core.Aura

	Judgement *core.Spell

	// The seal the paladin is under, and the ranks it has cast so far this iteration.
	currentSeal *sealConfig

	// The judgement effects this paladin can put on enemies; its melee strikes refresh them.
	JudgementAuras []core.AuraArray

	// Twist of Light: one Echo per seal that leaves one, keyed by the Echo's spell id.
	echoes map[int32]*sealEcho

	// Consecrated Ground: the targets the talent's bonus is active on, one aura per enemy.
	consecratedGroundAuras core.AuraArray

	// Light's Vigil: one entry per rank, so Holy Shock can find the vigil it consumes.
	lightsVigils []*lightsVigil

	// What gear adds to numbers the spells read as they register. Item effects and set bonuses
	// apply before Initialize, so the spells pick these up.
	sealOfTheCrusaderBonusAttackPower float64
	judgementOfTheCrusaderBonus       float64
	flashOfLightBonusHealing          float64
	holyShieldBlockValueMultiplier    float64
	forbearanceReduction              time.Duration

	// Timers shared by the ranks of one ability.
	judgementTimer     *core.Timer
	holyStrikeTimer    *core.Timer
	consecrationTimer  *core.Timer
	exorcismTimer      *core.Timer
	hammerOfWrathTimer *core.Timer
	holyWrathTimer     *core.Timer
	holyShockTimer     *core.Timer
	holyShieldTimer    *core.Timer
	lightsVigilTimer   *core.Timer
}

// Implemented by each Paladin spec.
type PaladinAgent interface {
	GetPaladin() *Paladin
}

func (paladin *Paladin) GetCharacter() *core.Character {
	return &paladin.Character
}

func (paladin *Paladin) GetPaladin() *Paladin {
	return paladin
}

func (paladin *Paladin) AddRaidBuffs(_ *proto.RaidBuffs) {
}

func (paladin *Paladin) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (paladin *Paladin) Initialize() {
	paladin.registerForbearance()
	paladin.registerRighteousFury()

	paladin.registerJudgement()
	paladin.registerSeals()
	paladin.registerAuras()

	HolyStrikeRankMap.RegisterAll(paladin.registerHolyStrike)
	ConsecrationRankMap.RegisterAll(paladin.registerConsecration)
	ExorcismRankMap.RegisterAll(paladin.registerExorcism)
	HammerOfWrathRankMap.RegisterAll(paladin.registerHammerOfWrath)
	HolyWrathRankMap.RegisterAll(paladin.registerHolyWrath)

	HolyLightRankMap.RegisterAll(paladin.registerHolyLight)
	FlashOfLightRankMap.RegisterAll(paladin.registerFlashOfLight)
	LayOnHandsRankMap.RegisterAll(paladin.registerLayOnHands)
}

func (paladin *Paladin) Reset(_ *core.Simulation) {
	paladin.currentSeal = nil
	for _, echo := range paladin.echoes {
		echo.seal = nil
	}
}

func (paladin *Paladin) OnEncounterStart(_ *core.Simulation) {
}

func (paladin *Paladin) GetMainHandType() proto.HandType {
	mh := paladin.GetMHWeapon()

	if mh != nil && (mh.HandType == proto.HandType_HandTypeTwoHand) {
		return proto.HandType_HandTypeTwoHand
	}

	return proto.HandType_HandTypeOneHand
}

func (paladin *Paladin) sharedTimer(timer **core.Timer) *core.Timer {
	if *timer == nil {
		*timer = paladin.NewTimer()
	}
	return *timer
}

// The cost a row states: a flat number, or a share of base mana for the spells the client prices
// that way (Judgement, Righteous Fury, Seal of Justice).
func manaCost(row shared.SpellData) core.ManaCostOptions {
	if row.PowerCostPct > 0 {
		return core.ManaCostOptions{BaseCostPercent: row.PowerCostPct}
	}
	return core.ManaCostOptions{FlatCost: row.Cost}
}

// The effect at the client's EffectIndex, for the effects Effect(aura, misc) cannot name: the
// weapon-damage effects carry no aura.
func effectAt(row shared.SpellData, index int32) shared.SpellDataEffect {
	for _, e := range row.Effects {
		if e.Index == index {
			return e
		}
	}
	panic("spell has no effect at the index")
}

func NewPaladin(character *core.Character, talentsStr string, _ *proto.PaladinOptions) *Paladin {
	paladin := &Paladin{
		Character: *character,
		Talents:   &proto.PaladinTalents{},

		holyShieldBlockValueMultiplier: 1,
	}

	core.FillTalentsProto(paladin.Talents.ProtoReflect(), talentsStr, TalentTreeSizes)

	paladin.PseudoStats.CanParry = true
	paladin.PseudoStats.BaseDodgeChance += 0.0065
	paladin.PseudoStats.BaseParryChance += 0.05
	paladin.PseudoStats.BaseBlockChance += 0.05

	paladin.EnableManaBar()

	paladin.EnableAutoAttacks(paladin, core.AutoAttackOptions{
		MainHand:       paladin.WeaponFromMainHand(),
		AutoSwingMelee: true,
	})

	paladin.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	paladin.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	paladin.AddStatDependency(stats.Agility, stats.DodgeRating, 1/25.0*core.DodgeRatingPerDodgePercent)
	paladin.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	return paladin
}
