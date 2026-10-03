package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{17, 19, 16}

type Warlock struct {
	core.Character
	Talents *proto.WarlockTalents
	Options *proto.WarlockOptions

	// Base Spells
	Corruption  *core.Spell
	DrainLife   *core.Spell
	Hellfire    *core.Spell
	Immolate    *core.Spell
	Incinerate  *core.Spell
	SearingPain *core.Spell
	ShadowBolt  *core.Spell
	Soulfire    *core.Spell

	LifeTap *core.Spell

	// Curses
	CurseOfAgony             *core.Spell
	CurseOfDoom              *core.Spell
	CurseOfElements          *core.Spell
	CurseOfElementsAuras     core.AuraArray
	CurseOfRecklessness      *core.Spell
	CurseOfRecklessnessAuras core.AuraArray

	// Talent Tree Spells
	AmplifyCurse *core.Spell
	Conflagrate  *core.Spell
	Shadowburn   *core.Spell
	SiphonLife   *core.Spell

	// Auras
	AmplifyCurseAura       *core.Aura
	NightfallProcAura      *core.Aura
	ImpShadowboltAura      *core.Aura
	ShadowEmbraceAura      *core.Aura
	DemonicKnowledgeAura   *core.Aura
	MasterDemonologistAura *core.Aura

	// Pets
	ActivePet  *WarlockPet
	Felhunter  *WarlockPet
	Imp        *WarlockPet
	Succubus   *WarlockPet
	Voidwalker *WarlockPet

	// Armors
	DemonArmor *core.Aura

	serviceTimer *core.Timer

	DemonicKnowledgeDep   *stats.StatDependency
	DemonicKnowledgeBonus float64

	currentActiveCurse *core.Spell

	CorruptionTickBaseDamage float64
	ImmolateTickBaseDamage   float64
	T5_4PC_Multiplier        map[int32]map[*core.Spell]float64
}

func (warlock *Warlock) GetCharacter() *core.Character {
	return &warlock.Character
}

func (warlock *Warlock) GetWarlock() *Warlock {
	return warlock
}

func RegisterWarlock() {
	core.RegisterAgentFactory(
		proto.Player_Warlock{},
		proto.Spec_SpecWarlock,
		func(character *core.Character, options *proto.Player, raid *proto.Raid) core.Agent {
			return NewWarlock(character, options, options.GetWarlock().Options.ClassOptions, raid)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_Warlock)
			if !ok {
				panic("Invalid spec value for Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

func (warlock *Warlock) Initialize() {

	// Curses
	warlock.registerCurseOfElements()
	warlock.registerCurseOfDoom()
	warlock.registerCurseOfAgony()
	warlock.registerCurseOfRecklessness()

	warlock.registerCorruption()
	warlock.registerDeathCoil()
	warlock.registerDrainLife()
	warlock.registerHellfire()
	warlock.registerImmolate()
	warlock.registerIncinerate()
	warlock.registerLifeTap()
	warlock.registerShadowBolt()
	warlock.registerSearingPain()
	warlock.registerSiphonLifeSpell()
	warlock.registerSoulfire()

	warlock.registerArmors()

	warlock.PseudoStats.SelfHealingMultiplier = 1.0
}

func (warlock *Warlock) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {

}

func (warlock *Warlock) AddPartyBuffs(partyBuffs *proto.PartyBuffs) {

}

func (warlock *Warlock) Reset(sim *core.Simulation) {
}

func (warlock *Warlock) OnEncounterStart(sim *core.Simulation) {}

func NewWarlock(character *core.Character, options *proto.Player, warlockOptions *proto.WarlockOptions, raid *proto.Raid) *Warlock {
	warlock := &Warlock{
		Character: *character,
		Talents:   &proto.WarlockTalents{},
		Options:   warlockOptions,
	}

	core.FillTalentsProto(warlock.Talents.ProtoReflect(), options.TalentsString, TalentTreeSizes)

	if raid.Debuffs != nil {
		switch warlock.Options.CurseOptions {
		case proto.WarlockOptions_Elements:
			raid.Debuffs.CurseOfElements = false
		case proto.WarlockOptions_Recklessness:
			raid.Debuffs.CurseOfRecklessness = false
		}
	}

	warlock.EnableManaBar()
	warlock.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	warlock.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[character.Class])

	if !warlock.Options.SacrificeSummon {
		warlock.registerPets()
	}

	warlock.T5_4PC_Multiplier = make(map[int32]map[*core.Spell]float64)

	return warlock
}

func (warlock *Warlock) AfflictionCount(target *core.Unit) float64 {
	return float64(len(target.GetAurasWithTag("Affliction")))
}

func (warlock *Warlock) DeactivateOtherCurses(sim *core.Simulation, newCurse *core.Spell, target *core.Unit) {
	if warlock.currentActiveCurse != nil {
		if warlock.currentActiveCurse.Dot(target) != nil {
			warlock.currentActiveCurse.Dot(target).Deactivate(sim)
		}
		if warlock.currentActiveCurse.RelatedAuraArrays != nil {
			for _, auraArray := range warlock.currentActiveCurse.RelatedAuraArrays {
				auraArray.Get(target).Deactivate(sim)
			}
		}
	}

	warlock.currentActiveCurse = newCurse
}

// Agent is a generic way to access underlying warlock on any of the agents.
type WarlockAgent interface {
	GetWarlock() *Warlock
}

const (
	WarlockSpellFlagNone    int64 = 0
	WarlockSpellConflagrate int64 = 1 << iota
	WarlockSpellShadowBolt
	WarlockSpellImmolate
	WarlockSpellImmolateDot
	WarlockSpellIncinerate
	WarlockSpellSoulFire
	WarlockSpellShadowBurn
	WarlockSpellLifeTap
	WarlockSpellCorruption
	WarlockSpellCurseOfAgony
	WarlockSpellCurseOfElements
	WarlockSpellDrainLife
	WarlockSpellHellfire
	WarlockSpellImmolationAura
	WarlockSpellSearingPain
	WarlockSpellSummonDoomguard
	WarlockSpellDoomguardDoomBolt
	WarlockSpellSummonImp
	WarlockSpellImpFireBolt
	WarlockSpellSummonFelhunter
	WarlockSpellFelHunterShadowBite
	WarlockSpellSummonSuccubus
	WarlockSpellSuccubusLashOfPain
	WarlockSpellVoidwalkerTorment
	WarlockSpellSummonInfernal
	WarlockSpellRainOfFire
	WarlockSpellCurseOfDoom
	WarlockSpellCurseOfRecklessness
	WarlockSpellCurseOfWeakness
	WarlockSpellSiphonLife
	WarlockSpellDrainSoul
	WarlockSpellDeathCoil
	WarlockSpellAll int64 = 1<<iota - 1

	WarlockShadowDamage = WarlockSpellCorruption | WarlockSpellDrainLife | WarlockSpellCurseOfAgony |
		WarlockSpellShadowBolt | WarlockSpellShadowBurn | WarlockSpellSiphonLife | WarlockSpellDeathCoil

	WarlockPeriodicShadowDamage = WarlockSpellCorruption |
		WarlockSpellDrainLife | WarlockSpellCurseOfAgony

	WarlockFireDamage = WarlockSpellConflagrate | WarlockSpellImmolate | WarlockSpellIncinerate | WarlockSpellSoulFire |
		WarlockSpellSearingPain | WarlockSpellImmolateDot | WarlockSpellShadowBurn

	WarlockDoT = WarlockSpellCorruption |
		WarlockSpellDrainLife | WarlockSpellCurseOfAgony | WarlockSpellImmolateDot

	WarlockSummonSpells = WarlockSpellSummonImp | WarlockSpellSummonSuccubus | WarlockSpellSummonFelhunter

	WarlockAllSummons = WarlockSummonSpells | WarlockSpellSummonInfernal | WarlockSpellSummonDoomguard

	WarlockContagionSpells = WarlockSpellCurseOfAgony | WarlockSpellCorruption

	WarlockShadowEmbraceSpells = WarlockSpellCorruption | WarlockSpellCurseOfAgony | WarlockSpellSiphonLife

	WarlockCurses = WarlockSpellCurseOfAgony | WarlockSpellCurseOfDoom | WarlockSpellCurseOfElements | WarlockSpellCurseOfRecklessness

	WarlockSoulLeechSpells = WarlockSpellShadowBolt | WarlockSpellShadowBurn | WarlockSpellSoulFire |
		WarlockSpellIncinerate | WarlockSpellSearingPain | WarlockSpellConflagrate

	WarlockAfflictionSpells = WarlockSpellCorruption | WarlockSpellCurseOfAgony | WarlockSpellCurseOfDoom | WarlockSpellCurseOfRecklessness | WarlockSpellCurseOfElements |
		WarlockSpellDrainLife | WarlockSpellDeathCoil

	WarlockDemonologySpells = WarlockAllSummons

	WarlockDestructionSpells = WarlockSpellHellfire | WarlockSpellImmolate | WarlockSpellIncinerate | WarlockSpellRainOfFire | WarlockSpellSearingPain |
		WarlockSpellShadowBolt | WarlockSpellSoulFire | WarlockSpellConflagrate | WarlockSpellShadowBurn
)

// Called to handle custom resources
type WarlockSpellCastedCallback func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation)
