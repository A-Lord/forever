package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type BaseStatsKey struct {
	Race  proto.Race
	Class proto.Class
}

var BaseStats = map[BaseStatsKey]stats.Stats{}

// TODO: These are LEVEL 70 attributes, but CharacterLevel is now 60. Changing the
// level constant does not rescale them -- unlike the ratings/crit/mana in
// base_stats_auto_gen.go, nothing regenerates these two maps, so a level-60
// character currently runs on level-70 str/agi/sta/int/spi. Needs a level-60
// source. Until then every sim number that depends on base attributes is wrong,
// and the goldens bake that in.
//
// The client is NOT that source, verified against build 1.60.1.69913 by extracting
// every candidate table: CharBaseInfo is race x class validity only; ChrClasses has
// AttackPowerPerStrength/Agility but no attributes; ChrRaces has no stat columns;
// RaceStat (new in 1.60.1) is one row per race and every value is 0; and
// PlayerExpectedStat carries BaseMana/CritPerAgility/SpellCritPerIntellect but no
// attributes. So these two maps have to stay log-fitted -- Classic-era WCL combatant
// info fitted the same way as the rows below, or a server emulator's player_levelstats.
//
// ClassBaseStats + RaceOffsets hold TRUE pre-racial base attributes: the
// multiplier racials (The Human Spirit ×1.05 spirit, applied via MultiplyStat
// in racials.go) are NOT included here.
// A naked character sheet shows floor(base × racial), e.g. human paladin
// spirit 89 shows as 93; multipliers (racial, Kings, %-stat talents) stack
// multiplicatively on the unfloored value with a single floor at the end.
//
// The game keeps one attribute row per race and class, but that table is a
// class row plus a race offset that is the same for every class, so the two
// maps below reproduce it exactly. Values come from WCL TBC-anniversary
// combatant info (Hyjal, 2026-09-18/19: 256 players across seven classes,
// each fitted through ComputeStats with the race, buffs, talents and
// consumables enumerated; warrior and rogue rows are unchanged). The hunter
// row was previously a wowhead-era guess; 38 hunters pin it, with strength
// verified on players without Strength of Earth or Kings.

// Base Spell Crit is calculated by
//   1. Take as-shown value (troll shaman have 3.5%)
//   2. Calculate the bonus from int (for troll shaman that would be 104/78.1=1.331% crit)
//   3. Subtract as-shown from int bouns (3.5-1.331=2.169)
//   4. 2.169*22.08 (rating per crit percent) = 47.89 crit rating.
//
// TODO: the 22.08 in step 4 is the level-70 crit rating per percent. At level 60
// it is 14 (see SpellCritRatingPerCritPercent in base_stats_auto_gen.go), so this
// worked example no longer reproduces the numbers above it.

// Base mana can be looked up here: https://wowwiki-archive.fandom.com/wiki/Base_mana

// These are also scattered in various dbc/casc files,
// `octbasempbyclass.txt`, `combatratings.txt`, `chancetospellcritbase.txt`, etc.

var RaceOffsets = map[proto.Race]stats.Stats{
	proto.Race_RaceUnknown: stats.Stats{},
	proto.Race_RaceHuman:   stats.Stats{},
	proto.Race_RaceOrc: {
		stats.Agility:   -3,
		stats.Strength:  3,
		stats.Intellect: -3,
		stats.Spirit:    3,
		stats.Stamina:   2,
	},
	proto.Race_RaceDwarf: {
		stats.Agility:   -4,
		stats.Strength:  2,
		stats.Intellect: -1,
		stats.Spirit:    -1,
		stats.Stamina:   3,
	},
	proto.Race_RaceNightElf: {
		stats.Agility:   5,
		stats.Strength:  -3,
		stats.Intellect: 0,
		stats.Spirit:    0,
		stats.Stamina:   -1,
	},
	proto.Race_RaceUndead: {
		stats.Agility:   -2,
		stats.Strength:  -1,
		stats.Intellect: -2,
		stats.Spirit:    5,
		stats.Stamina:   1,
	},
	proto.Race_RaceTauren: {
		stats.Agility:   -5,
		stats.Strength:  5,
		stats.Intellect: -5,
		stats.Spirit:    2,
		stats.Stamina:   2,
	},
	proto.Race_RaceGnome: {
		stats.Agility:   3,
		stats.Strength:  -5,
		stats.Intellect: 3,
		stats.Spirit:    0,
		stats.Stamina:   -1,
	},
	proto.Race_RaceTroll: {
		stats.Agility:   2,
		stats.Strength:  1,
		stats.Intellect: -4,
		stats.Spirit:    1,
		stats.Stamina:   1,
	},
	// Read from level 1 naked character sheets (2026-09-22): four Human/High
	// Order pairs (warrior, hunter, mage, rogue) give the same delta, and a
	// Windshaper warrior matches the High Order one, so both variants share it.
	proto.Race_RaceHighOrderSkyborne: {
		stats.Agility:   1,
		stats.Strength:  -1,
		stats.Intellect: 1,
		stats.Spirit:    0,
		stats.Stamina:   -1,
	},
	proto.Race_RaceWindshaperSkyborne: {
		stats.Agility:   1,
		stats.Strength:  -1,
		stats.Intellect: 1,
		stats.Spirit:    0,
		stats.Stamina:   -1,
	},
}

var ClassBaseStats = map[proto.Class]stats.Stats{
	proto.Class_ClassUnknown: {},
	proto.Class_ClassWarrior: {
		stats.Health:      4264,
		stats.Agility:     96,
		stats.Strength:    145,
		stats.Intellect:   33,
		stats.Spirit:      51,
		stats.Stamina:     133,
		stats.AttackPower: float64(CharacterLevel)*3.0 - 20,
	},
	proto.Class_ClassPaladin: {
		stats.Health:      3197,
		stats.Agility:     77,
		stats.Strength:    126,
		stats.Intellect:   83,
		stats.Spirit:      89,
		stats.Stamina:     120,
		stats.AttackPower: float64(CharacterLevel)*3.0 - 20,
	},
	proto.Class_ClassHunter: {
		stats.Health:            3388,
		stats.Agility:           151,
		stats.Strength:          64,
		stats.Intellect:         77,
		stats.Spirit:            83,
		stats.Stamina:           108,
		stats.AttackPower:       float64(CharacterLevel)*2.0 - 20,
		stats.RangedAttackPower: float64(CharacterLevel)*2.0 - 20,
	},
	proto.Class_ClassRogue: {
		stats.Health:      3524,
		stats.Agility:     158,
		stats.Strength:    95,
		stats.Intellect:   39,
		stats.Spirit:      58,
		stats.Stamina:     89,
		stats.AttackPower: float64(CharacterLevel)*2.0 - 20,
	},
	proto.Class_ClassPriest: {
		stats.Health:    3211,
		stats.Agility:   45,
		stats.Strength:  39,
		stats.Intellect: 145,
		stats.Spirit:    151,
		stats.Stamina:   58,
	},
	proto.Class_ClassShaman: {
		stats.Health:      2979,
		stats.Agility:     64,
		stats.Strength:    102,
		stats.Intellect:   108,
		stats.Spirit:      120,
		stats.Stamina:     114,
		stats.AttackPower: float64(CharacterLevel) * 2.0,
	},
	proto.Class_ClassMage: {
		stats.Health:    3213,
		stats.Agility:   39,
		stats.Strength:  33,
		stats.Intellect: 151,
		stats.Spirit:    145,
		stats.Stamina:   51,
	},
	proto.Class_ClassWarlock: {
		stats.Health:      3310,
		stats.Agility:     58,
		stats.Strength:    51,
		stats.Intellect:   133,
		stats.Spirit:      139,
		stats.Stamina:     76,
		stats.AttackPower: -10,
	},
	proto.Class_ClassDruid: {
		stats.Health:      3434,
		stats.Agility:     70,
		stats.Strength:    76,
		stats.Intellect:   120,
		stats.Spirit:      133,
		stats.Stamina:     83,
		stats.AttackPower: -20,
	},
}

// The LEVEL 60 rows of GameTables/SpellScaling.txt. Only tools/tooltip reads this
// (dbc_data_provider.go:249), so the sim is unaffected either way, but tooltip spell
// values used to be computed off the level 90 row -- a MoP-port leftover that was never
// updated for TBC and was wrong by two expansions at level 60.
//
// TODO: these numbers came from a MoP-era copy of that gametable. The checked-in
// assets/db_inputs/basestats/SpellScaling.txt is now the beta's own extraction, and
// every class column in it is 0 at every level, so there is no build-native
// confirmation of these numbers and none is currently obtainable.
var ClassBaseScaling = map[proto.Class]float64{
	proto.Class_ClassUnknown: 49.000000,
	proto.Class_ClassWarrior: 491.949980,
	proto.Class_ClassPaladin: 332.962490,
	proto.Class_ClassHunter:  355.055050,
	proto.Class_ClassRogue:   532.945800,
	proto.Class_ClassPriest:  336.625000,
	proto.Class_ClassShaman:  251.970830,
	proto.Class_ClassMage:    366.620820,
	proto.Class_ClassWarlock: 308.962490,
	proto.Class_ClassDruid:   282.633330,
}

func AddBaseStatsCombo(r proto.Race, c proto.Class) {
	BaseStats[BaseStatsKey{Race: r, Class: c}] = ClassBaseStats[c].Add(RaceOffsets[r]).Add(ExtraClassBaseStats[c])
}

func init() {
	for class, races := range ClassRaceCapabilities {
		for _, race := range races {
			AddBaseStatsCombo(race, class)
		}
	}
}
