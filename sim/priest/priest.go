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
	DevouringPlagueRankMap.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerDevouringPlagueSpell(rank)
	})
	priest.registerShadowfiendSpell()
	spellData.InnerFire.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerInnerFire(rank)
	})
	spellData.Fade.Each(func(_ int32, rank *spelldata.Spell) {
		priest.registerFade(rank)
	})

	if priest.Race == proto.Race_RaceNightElf {
		StarshardsRankMap.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerStarshardsSpell(rank)
		})
		spellData.ElunesGrace.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerElunesGrace(rank)
		})
	}
	if priest.Race == proto.Race_RaceTroll {
		spellData.HexOfWeakness.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerHexOfWeakness(rank)
		})
		spellData.Shadowguard.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerShadowguard(rank)
		})
	}
	if priest.Race == proto.Race_RaceUndead {
		spellData.TouchOfWeakness.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerTouchOfWeakness(rank)
		})
		spellData.DarkSacrifice.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerDarkSacrifice(rank)
		})
	}
	if priest.Race == proto.Race_RaceDwarf {
		spellData.Chastise.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerChastise(rank)
		})
		spellData.DesperatePrayer.Each(func(_ int32, rank *spelldata.Spell) {
			priest.registerDesperatePrayer(rank)
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
	PriestSpellDesperatePrayer

	// TODO: Forever abilities the sim does not model yet; see the stub file named for each.
	PriestSpellConfoundingFlash
	PriestSpellContingencyPlan
	PriestSpellDivineGrace

	PriestShadowSpells = PriestSpellDevouringPlague |
		PriestSpellShadowWordDeath |
		PriestSpellShadowform |
		PriestSpellShadowWordPain |
		PriestSpellMindFlay |
		PriestSpellMindBlast |
		PriestSpellShadowFiend |
		PriestSpellVampiricEmbrace
)
