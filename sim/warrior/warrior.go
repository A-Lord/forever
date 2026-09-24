package warrior

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{17, 18, 18}

type WarriorInputs struct {
	DefaultShout  proto.WarriorShout
	DefaultStance proto.WarriorStance

	StartingRage   float64
	StanceSnapshot bool
	HasBsT2        bool
}

// What is left of the sim's own spell masks now that the client's spell class mask addresses the
// rest: the two Sweeping Strikes copies and the Whirlwind off-hand strike, which have no client row
// to be named by, and the six abilities a handler or a listener singles out by hand.
const (
	SpellMaskNone int64 = 0

	SpellMaskSweepingStrikesHit int64 = 1 << iota
	SpellMaskSweepingStrikesNormalizedHit
	SpellMaskWhirlwindOh

	SpellMaskExecute
	SpellMaskThunderClap
	SpellMaskWhirlwind
	SpellMaskShieldBash
	SpellMaskBloodthirst
	SpellMaskHamstring

	WarriorSpellLast
	WarriorSpellsAll = WarriorSpellLast<<1 - 1
)

// The client's SpellClassOptions for the registrations that do not resolve a row of their own: the
// stance spells, and the sim-only sub-spells that take their parent's flags, the client having no
// row for a Whirlwind off-hand strike or a Sweeping Strikes copy.
var (
	SpellFlagsBattleStance    = spellData.BattleStance.Highest().ClassFlags
	SpellFlagsBerserkerStance = spellData.BerserkerStance.Highest().ClassFlags
	SpellFlagsDefensiveStance = spellData.DefensiveStance.Highest().ClassFlags

	SpellFlagsSweepingStrikes = spellData.SweepingStrikes.Highest().ClassFlags
	SpellFlagsWhirlwind       = spellData.Whirlwind.Highest().ClassFlags
)

type Warrior struct {
	core.Character

	ClassSpellScaling float64

	Talents *proto.WarriorTalents

	WarriorInputs

	// Current state
	thunderClapEffectBonus float64

	BattleShout       *core.Spell
	DemoralizingShout *core.Spell
	BattleStance      *core.Spell
	DefensiveStance   *core.Spell
	BerserkerStance   *core.Spell

	Rend                            *core.Spell
	DeepWounds                      *core.Spell
	MortalStrike                    *core.Spell
	SweepingStrikesNormalizedAttack *core.Spell

	HeroicStrike      *core.Spell
	Cleave            *core.Spell
	MockingBlow       *core.Spell
	ChallengingShout  *core.Spell
	IntimidatingShout *core.Spell
	Disarm            *core.Spell
	Taunt             *core.Spell
	VictoryRush       *core.Spell

	EnrageAura *core.Aura

	SweepingStrikesAura *core.Aura
	OverpowerAura       *core.Aura

	DemoralizingShoutAuras core.AuraArray
	SunderArmorAuras       core.AuraArray
}

func (warrior *Warrior) GetCharacter() *core.Character {
	return &warrior.Character
}

func (warrior *Warrior) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
}

func (warrior *Warrior) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (warrior *Warrior) Initialize() {
	warrior.registerRecklessness()
	warrior.registerShieldWall()
	warrior.registerRetaliation()

	warrior.registerBerserkerRage()
	warrior.registerBloodrage()
	warrior.registerCharge()
	warrior.registerIntercept()
	warrior.registerPummel()
	warrior.registerHamstring()
	warrior.registerDisarm()
	warrior.registerTaunt()

	warrior.registerRend()
	warrior.registerSunderArmor()
	warrior.registerHeroicStrike()
	warrior.registerCleave()
	warrior.registerOverpower()
	warrior.registerSlam()
	warrior.registerWhirlwind()
	warrior.registerExecute()
	warrior.registerThunderClap()
	warrior.registerRevenge()
	warrior.registerShieldBlock()
	warrior.registerShieldBash()
	warrior.registerMockingBlow()
	warrior.registerVictoryRush()

	warrior.registerStances()
	warrior.registerBattleShout()
	warrior.registerDemoralizingShout()
	warrior.registerChallengingShout()
	warrior.registerIntimidatingShout()
}

func (warrior *Warrior) Reset(_ *core.Simulation) {
	switch warrior.DefaultStance {
	case proto.WarriorStance_WarriorStanceBattle:
		warrior.ShapeshiftForm = battleStanceRank.ShapeshiftForm()
	case proto.WarriorStance_WarriorStanceDefensive:
		warrior.ShapeshiftForm = defensiveStanceRank.ShapeshiftForm()
	case proto.WarriorStance_WarriorStanceBerserker:
		warrior.ShapeshiftForm = berserkerStanceRank.ShapeshiftForm()
	}
}

func (warrior *Warrior) OnEncounterStart(sim *core.Simulation) {}

func (warrior *Warrior) GetMainHandType() proto.HandType {
	mh := warrior.GetMHWeapon()

	if mh != nil && (mh.HandType == proto.HandType_HandTypeTwoHand) {
		return proto.HandType_HandTypeTwoHand
	}

	return proto.HandType_HandTypeOneHand
}

func NewWarrior(character *core.Character, options *proto.WarriorOptions, talents string, inputs WarriorInputs) *Warrior {
	warrior := &Warrior{
		Character:     *character,
		Talents:       &proto.WarriorTalents{},
		WarriorInputs: inputs,
	}
	core.FillTalentsProto(warrior.Talents.ProtoReflect(), talents, TalentTreeSizes)

	warrior.EnableRageBar(core.RageBarOptions{
		MaxRage:            100 + spellData.BoundlessRage.TenthsAt(warrior.Talents.BoundlessRage),
		BaseRageMultiplier: 1,
		StartingRage:       inputs.StartingRage,
	})

	warrior.EnableAutoAttacks(warrior, core.AutoAttackOptions{
		MainHand:       warrior.WeaponFromMainHand(),
		OffHand:        warrior.WeaponFromOffHand(),
		AutoSwingMelee: true,
	})

	warrior.PseudoStats.CanParry = true
	// TODO: In-game testing required
	warrior.PseudoStats.BaseDodgeChance += 0.0075
	warrior.PseudoStats.BaseParryChance += 0.05
	warrior.PseudoStats.BaseBlockChance += 0.05

	warrior.AddStatDependency(stats.Strength, stats.AttackPower, 2)
	warrior.AddStatDependency(stats.Strength, stats.BlockValue, 1/20.0)
	warrior.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])
	warrior.AddStatDependency(stats.Agility, stats.DodgeRating, 1/30.0*core.DodgeRatingPerDodgePercent)
	warrior.AddStatDependency(stats.BonusArmor, stats.Armor, 1)

	return warrior
}

func (warrior *Warrior) CastNormalizedSweepingStrikesAttack(results core.SpellResultSlice, sim *core.Simulation) {
	if warrior.SweepingStrikesAura != nil && warrior.SweepingStrikesAura.IsActive() {
		for _, result := range results {
			if result.Landed() {
				warrior.SweepingStrikesNormalizedAttack.Cast(sim, warrior.Env.NextActiveTargetUnit(result.Target))
				warrior.SweepingStrikesAura.RemoveStack(sim)
				break
			}
		}
	}
}

// Agent is a generic way to access underlying warrior on any of the agents.
type WarriorAgent interface {
	GetWarrior() *Warrior
}

// The recovery the ability waits out. A warrior ability states it on a shared category - Bloodthirst
// and Mortal Strike both run off category 971 - and leaves its own column at zero, so the cooldown
// is whichever of the two the client filled in.
//
// The category is a timer as well as a number, and spelldata.SpellConfig puts every spell naming one
// on the unit's timer for it: Revenge and Overpower share category 65, Shield Bash and Pummel share
// 88, and Mortal Strike, Bloodthirst and Shield Slam share 971.
func cooldownOf(s *spelldata.Spell) time.Duration {
	return max(s.Cooldown(), s.CategoryCooldown())
}
