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
	DrainSoul   *core.Spell
	Hellfire    *core.Spell
	Immolate    *core.Spell
	Incinerate  *core.Spell
	SearingPain *core.Spell
	ShadowBolt  *core.Spell
	Soulfire    *core.Spell

	LifeTap *core.Spell

	// Curses
	CurseOfElements          *core.Spell
	CurseOfElementsAuras     core.AuraArray
	CurseOfRecklessness      *core.Spell
	CurseOfRecklessnessAuras core.AuraArray
	Curses                   []*core.Spell

	// Banes
	BaneOfAgony *core.Spell
	BaneOfDoom  *core.Spell
	Banes       []*core.Spell

	// Talent Tree Spells
	AmplifyCurse *core.Spell
	Conflagrate  *core.Spell
	Shadowburn   *core.Spell
	SiphonLife   *core.Spell
	Wrack        *core.Spell

	// Auras
	AmplifyCurseAura       *core.Aura
	AmplifyCurseMultiplier float64
	NightfallProcAura      *core.Aura
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
	warlock.registerBaneOfDoom()
	warlock.registerBaneOfAgony()
	warlock.registerCurseOfRecklessness()

	warlock.registerCorruption()
	warlock.registerDeathCoil()
	warlock.registerDrainLife()
	warlock.registerDrainSoul()
	warlock.registerHellfire()
	warlock.registerImmolate()
	warlock.registerIncinerate()
	warlock.registerLifeTap()
	warlock.registerShadowBolt()
	warlock.registerSearingPain()

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

	if !(warlock.Options.SacrificeSummon && warlock.Talents.DemonicSacrifice) {
		warlock.registerPets()
	}

	warlock.T5_4PC_Multiplier = make(map[int32]map[*core.Spell]float64)

	return warlock
}

// Returns the number of active affliction effects on the target.
func (warlock *Warlock) AfflictionCount(target *core.Unit) float64 {
	count := 0.0
	for _, aura := range target.GetAurasWithTag("Affliction") {
		if aura.IsActive() {
			count++
		}
	}
	return count
}

// Only one Bane per Warlock can be active on any one target, so a Bane landing takes the others off it.
func (warlock *Warlock) deactivateOtherBanes(sim *core.Simulation, bane *core.Spell, target *core.Unit) {
	deactivateOthersOnTarget(sim, warlock.Banes, bane, target)
}

// Only one Curse per Warlock can be active on any one target, so a Curse landing takes the others off it.
func (warlock *Warlock) deactivateOtherCurses(sim *core.Simulation, curse *core.Spell, target *core.Unit) {
	deactivateOthersOnTarget(sim, warlock.Curses, curse, target)
}

// Takes every spell in the provided group off of the target, except the landing spell.
func deactivateOthersOnTarget(sim *core.Simulation, group []*core.Spell, landing *core.Spell, target *core.Unit) {
	for _, other := range group {
		if other == landing {
			continue
		}
		if dot := other.Dot(target); dot != nil {
			dot.Deactivate(sim)
		}
		for _, auras := range other.RelatedAuraArrays {
			if aura := auras.Get(target); aura != nil {
				aura.Deactivate(sim)
			}
		}
	}
}

// Slots in a target's per-caster damage multipliers (core.EnableDamageDoneByCaster / AttachDDBC), one per
// warlock debuff that raises only this warlock's damage against the target. Each needs its own slot so one
// debuff's expiry cannot clear another's, and every caller passes DDBC_Total so the slice is sized alike.
const (
	DDBC_ShadowVulnerability int = iota
	DDBC_Wrack
	DDBC_Total
)

// Scales a result already calculated - a ramp, a per-target bonus - and keeps its threat in step with
// the damage, which the calculation set from the damage before the scaling.
//
// Landed warlock dots tick through spelldata's TickOutcomeHitRolled: they tested hit once, when applied,
// so a tick rolls only the crit the row's Periodic Can Crit allows (the Forever beta log has no
// SPELL_PERIODIC_MISSED). Area effects that tick, Hellfire and Rain of Fire, roll hit on every tick.
func scaleResult(sim *core.Simulation, spell *core.Spell, result *core.SpellResult, multiplier float64) {
	result.Damage *= multiplier
	result.Threat = spell.ThreatFromDamage(sim, result.Outcome, result.Damage, spell.Unit.AttackTables[result.Target.UnitIndex])
}

// Combines the class flags of several spells into one set, for filters the data doesn't name the spells for.
// All of them must share a family, as a warlock's spells do.
func combineClassFlags(flags ...core.ClassFlags) core.ClassFlags {
	var combined core.ClassFlags
	for _, f := range flags {
		if combined.Family != 0 && f.Family != combined.Family {
			panic("combineClassFlags: spells from different families")
		}
		combined.Family = f.Family
		for i := range combined.Mask {
			combined.Mask[i] |= f.Mask[i]
		}
	}
	return combined
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
	WarlockSpellBaneOfAgony
	WarlockSpellBaneOfDoom
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
	WarlockSpellCurseOfRecklessness
	WarlockSpellCurseOfWeakness
	WarlockSpellSiphonLife
	WarlockSpellWrack
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

	WarlockBanes = WarlockSpellBaneOfAgony | WarlockSpellBaneOfDoom

	WarlockCurses = WarlockSpellCurseOfElements | WarlockSpellCurseOfRecklessness

	WarlockSoulLeechSpells = WarlockSpellShadowBolt | WarlockSpellShadowBurn | WarlockSpellSoulFire |
		WarlockSpellIncinerate | WarlockSpellSearingPain | WarlockSpellConflagrate

	WarlockAfflictionSpells = WarlockSpellCorruption | WarlockSpellCurseOfAgony | WarlockSpellCurseOfRecklessness | WarlockSpellCurseOfElements |
		WarlockSpellDrainLife | WarlockSpellDeathCoil | WarlockSpellSiphonLife | WarlockSpellWrack

	WarlockDemonologySpells = WarlockAllSummons

	WarlockDestructionSpells = WarlockSpellHellfire | WarlockSpellImmolate | WarlockSpellIncinerate | WarlockSpellRainOfFire | WarlockSpellSearingPain |
		WarlockSpellShadowBolt | WarlockSpellSoulFire | WarlockSpellConflagrate | WarlockSpellShadowBurn
)

// Called to handle custom resources
type WarlockSpellCastedCallback func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation)
