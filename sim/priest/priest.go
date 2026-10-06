package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var TalentTreeSizes = [3]int{18, 17, 18}

type Priest struct {
	core.Character
	SelfBuffs
	Talents *proto.PriestTalents

	Latency float64

	ShadowfiendAura *core.Aura
	ShadowfiendPet  *Shadowfiend

	Shadowfiend    *core.Spell
	InnerFocusAura *core.Aura

	VampiricEmbrace *core.Spell
}

type SelfBuffs struct {
	UseShadowfiend bool
	PreShadowform  bool
	Armor          proto.PriestOptions_Armor
}

func (priest *Priest) GetCharacter() *core.Character {
	return &priest.Character
}
func (priest *Priest) GetPriest() *Priest {
	return priest
}

func (priest *Priest) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (priest *Priest) Initialize() {
	HolyFireRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerHolyFireSpell(rank)
	})
	MindBlastRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerMindBlastSpell(rank)
	})
	ShadowWordPainRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerShadowWordPainSpell(rank)
	})
	ShadowWordDeathRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerShadowWordDeathSpell(rank)
	})
	SmiteRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerSmiteSpell(rank)
	})
	// TODO: Confirm in game that every race has it; the game files grant all six ranks to every race
	// (SkillLineAbility RaceMasks -1).
	DevouringPlagueRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerDevouringPlagueSpell(rank)
	})
	priest.registerShadowfiendSpell()

	if priest.Race == proto.Race_RaceNightElf {
		StarshardsRankMap.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerStarshardsSpell(rank)
		})
	}
	if priest.Race == proto.Race_RaceUndead {
		spellData.DarkSacrifice.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerDarkSacrifice(rank)
		})
	}
	if priest.Race == proto.Race_RaceDwarf {
		spellData.Chastise.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerChastise(rank)
		})
	}
}

func (priest *Priest) Reset(_ *core.Simulation) {

}

func (priest *Priest) OnEncounterStart(sim *core.Simulation) {
}

func New(char *core.Character, selfBuffs SelfBuffs, talents string) *Priest {
	priest := &Priest{
		Character: *char,
		SelfBuffs: selfBuffs,
		Talents:   &proto.PriestTalents{},
	}

	core.FillTalentsProto(priest.Talents.ProtoReflect(), talents, TalentTreeSizes)
	priest.EnableManaBar()
	priest.AddStatDependency(stats.Strength, stats.AttackPower, 1)
	if selfBuffs.Armor == proto.PriestOptions_InnerFire {
		// Inner Fire rank 7: +1580 armor. The charges never run out on a caster that is not hit.
		priest.AddStat(stats.Armor, 1580)
	}
	priest.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[char.Class])
	priest.ShadowfiendPet = priest.NewShadowfiend()

	return priest
}

// Agent is a generic way to access underlying priest on any of the agents.
type PriestAgent interface {
	GetPriest() *Priest
}

func NewPriest(character *core.Character, options *proto.Player) *Priest {
	classOptions := options.GetDpsPriest().GetOptions().GetClassOptions()
	selfBuffs := SelfBuffs{
		UseShadowfiend: true,
		PreShadowform:  classOptions.GetPreShadowform(),
		Armor:          classOptions.GetArmor(),
	}

	basePriest := New(character, selfBuffs, options.TalentsString)
	basePriest.Latency = float64(basePriest.ChannelClipDelay.Milliseconds())

	return basePriest
}

func RegisterPriest() {
	core.RegisterAgentFactory(
		proto.Player_DpsPriest{},
		proto.Spec_SpecDpsPriest,
		func(character *core.Character, options *proto.Player, _ *proto.Raid) core.Agent {
			return NewPriest(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_DpsPriest)
			if !ok {
				panic("Invalid spec value for Priest!")
			}
			player.Spec = playerSpec
		},
	)
}

const (
	PriestSpellFlagNone        int64 = 0
	PriestSpellDevouringPlague int64 = 1 << iota
	PriestSpellDevouringPlagueDoT
	PriestSpellDevouringPlagueHeal
	PriestSpellHolyNova
	PriestSpellHolyFire
	PriestSpellMindBlast
	PriestSpellMindFlay
	PriestSpellPowerInfusion
	PriestSpellStarshards
	PriestSpellShadowform
	PriestSpellShadowWordDeath
	PriestSpellShadowWordPain
	PriestSpellShadowFiend
	PriestSpellVampiricEmbrace
	PriestSpellFade
	PriestSpellSmite
	PriestSpellInnerFire
	PriestSpellChastise
	PriestSpellDarkSacrifice
	PriestSpellPenance

	// TODO: Forever abilities the sim does not model yet; see the stub file named for each.
	PriestSpellConfoundingFlash
	PriestSpellContingencyPlan
	PriestSpellDivineGrace

	// Priest Healing Spells for Shadowform (and eventual healing sim?)
	PriestSpellLesserHeal
	PriestSpellHeal
	PriestSpellGreaterHeal
	PriestSpellFlashHeal
	PriestSpellBindingHeal
	PriestSpellDesperatePrayer
	PriestSpellPrayerOfHealing
	PriestSpellPrayerOfMending
	PriestSpellRenew
	PriestSpellLightwell
	PriestSpellPowerWordShield

	PriestSpellCureDisease
	PriestSpellAbolishDisease
	PriestSpellDispelMagic
	PriestSpellFearWard

	PriestSpellLast
	PriestSpellsAll    = PriestSpellLast<<1 - 1
	PriestSpellDoT     = PriestSpellDevouringPlague | PriestSpellHolyFire | PriestSpellMindFlay | PriestSpellShadowWordPain | PriestSpellStarshards
	PriestSpellInstant = PriestSpellDevouringPlague |
		PriestSpellFade |
		PriestSpellHolyNova |
		PriestSpellPowerInfusion |
		PriestSpellShadowWordDeath |
		PriestSpellShadowWordPain |
		PriestSpellVampiricEmbrace |
		PriestSpellShadowFiend |
		PriestSpellStarshards |
		PriestSpellShadowform |
		PriestSpellPowerInfusion
	PriestShadowSpells = PriestSpellDevouringPlague |
		PriestSpellShadowWordDeath |
		PriestSpellShadowform |
		PriestSpellShadowWordPain |
		PriestSpellMindFlay |
		PriestSpellMindBlast |
		PriestSpellShadowFiend |
		PriestSpellVampiricEmbrace
	PriestHolySpells    = PriestSpellSmite | PriestSpellHolyFire | PriestSpellHolyNova
	PriestHealingSpells = PriestSpellLesserHeal |
		PriestSpellHeal |
		PriestSpellGreaterHeal |
		PriestSpellFlashHeal |
		PriestSpellBindingHeal |
		PriestSpellDesperatePrayer |
		PriestSpellPrayerOfHealing |
		PriestSpellPrayerOfMending |
		PriestSpellPenance |
		PriestSpellPowerWordShield |
		PriestSpellRenew |
		PriestSpellLightwell |
		PriestSpellPowerWordShield |
		PriestSpellCureDisease |
		PriestSpellAbolishDisease |
		PriestSpellDispelMagic |
		PriestSpellFearWard
)
