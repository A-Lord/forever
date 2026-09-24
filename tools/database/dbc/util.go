package dbc

import (
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func GetProfession(id int) proto.Profession {
	if profession, ok := MapProfessionIdToProfession[id]; ok {
		return profession
	}
	return 0
}

func GetRepLevel(minReputation int) proto.RepLevel {
	if repLevel, ok := MapMinReputationToRepLevel[minReputation]; ok {
		return repLevel
	}
	return proto.RepLevel_RepLevelUnknown
}
func NullFloat(arr []float64) []float64 {
	for _, v := range arr {
		if v > 0 {
			return arr
		}
	}

	return nil
}

// The playable races by ChrRaces.ID, whose bit in ItemSparse.AllowableRace is ID - 1.
var playableRaces = []struct {
	race proto.Race
	id   int
}{
	{proto.Race_RaceHuman, 1}, {proto.Race_RaceOrc, 2}, {proto.Race_RaceDwarf, 3}, {proto.Race_RaceNightElf, 4},
	{proto.Race_RaceUndead, 5}, {proto.Race_RaceTauren, 6}, {proto.Race_RaceGnome, 7}, {proto.Race_RaceTroll, 8},
	{proto.Race_RaceBloodElf, 10}, {proto.Race_RaceDraenei, 11},
}

// The playable races an AllowableRace mask names, or none where it names every one of them.
func GetRacesFromRaceMask(mask int64) []proto.Race {
	var result []proto.Race
	for _, r := range playableRaces {
		if mask&(1<<(r.id-1)) != 0 {
			result = append(result, r.race)
		}
	}
	if len(result) == len(playableRaces) {
		return nil
	}
	slices.Sort(result)
	return result
}

func GetClassesFromClassMask(mask int) []proto.Class {
	var result []proto.Class

	allClasses := (1 << len(Classes)) - 1
	if mask&allClasses == allClasses {
		return result
	}

	for _, class := range Classes {
		if mask&(1<<(class.ID-1)) != 0 {
			result = append(result, class.ProtoClass)
		}
	}
	slices.Sort(result)
	return result
}

func WriteGzipFile(filePath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", filePath, err)
	}
	// Create the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Create a gzip writer on top of the file writer
	gw := gzip.NewWriter(f)
	defer gw.Close()

	// Write the data to the gzip writer
	_, err = gw.Write(data)
	return err
}
func ReadGzipFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "gzip file",
			Reason:   err.Error(),
		}
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "gzip",
			Reason:   err.Error(),
		}
	}
	defer gzReader.Close()

	data, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, DataLoadError{
			Source:   filename,
			DataType: "decompression",
			Reason:   err.Error(),
		}
	}

	return data, nil
}

func processEnchantmentEffects(
	effects []int,
	effectArgs []int,
	effectPoints []int,
	outStats *stats.Stats,
	outPseudoStats []float64,
	addRanged bool,
) {
	for i, effect := range effects {
		switch effect {
		case ITEM_ENCHANTMENT_RESISTANCE:
			stat, match := MapResistanceToStat(effectArgs[i])
			if !match {
				continue
			}
			outStats[stat] = float64(effectPoints[i])
		case ITEM_ENCHANTMENT_STAT:
			mapped, success := MapBonusStatIndexToStats(effectArgs[i])
			if !success {
				continue
			}
			for _, stat := range mapped {
				outStats[stat] = float64(effectPoints[i])

				// If the bonus stat is attack power, copy it to ranged attack power
				if addRanged && stat == proto.Stat_StatAttackPower {
					outStats[proto.Stat_StatRangedAttackPower] = float64(effectPoints[i])
				}
			}
		case ITEM_ENCHANTMENT_EQUIP_SPELL: //Buff
			if outPseudoStats != nil {
				AddEquipSpellPseudoStats(outPseudoStats, effectArgs[i])
			}
			AddEquipSpellStats(outStats, effectArgs[i])
			spellEffects := dbcInstance.SpellEffects[effectArgs[i]]
			for _, spellEffect := range spellEffects {
				points := spellEffect.EffectBasePoints + spellEffect.EffectDieSides

				if spellEffect.EffectMiscValues[0] == -1 &&
					spellEffect.EffectType == dbcenums.E_APPLY_AURA &&
					spellEffect.EffectAura == dbcenums.A_MOD_STAT {
					// Apply bonus to all stats
					outStats[proto.Stat_StatAgility] += float64(points)
					outStats[proto.Stat_StatIntellect] += float64(points)
					outStats[proto.Stat_StatSpirit] += float64(points)
					outStats[proto.Stat_StatStamina] += float64(points)
					outStats[proto.Stat_StatStrength] += float64(points)
					continue
				}
				if spellEffect.EffectType == dbcenums.E_APPLY_AURA && spellEffect.EffectAura == dbcenums.A_MOD_STAT {
					stat, ok := MapMainStatToStat(spellEffect.EffectMiscValues[0])
					if !ok {
						continue
					}
					outStats[stat] += float64(points)
				} else if spellEffect.EffectType == dbcenums.E_APPLY_AURA && spellEffect.EffectAura == dbcenums.A_MOD_RESISTANCE && (SpellSchool(spellEffect.EffectMiscValues[0]) == ALL_SPELL_DAMAGE || SpellSchool(spellEffect.EffectMiscValues[0]) == SPELL_PENETRATION) {
					outStats[proto.Stat_StatArcaneResistance] += float64(points)
					outStats[proto.Stat_StatFireResistance] += float64(points)
					outStats[proto.Stat_StatFrostResistance] += float64(points)
					outStats[proto.Stat_StatNatureResistance] += float64(points)
					outStats[proto.Stat_StatShadowResistance] += float64(points)
				} else {
					stat := ConvertEffectAuraToStatIndex(spellEffect.EffectAura, spellEffect.EffectMiscValues[0])
					if stat >= 0 || stat == -2 {
						value := float64(points)
						if stat == proto.Stat_StatArmorPenetration || stat == proto.Stat_StatSpellPiercing {
							// Make sure it's not Feral AP
							if strings.Contains(dbcInstance.Spells[spellEffect.SpellID].Description, "forms only") {
								stat = proto.Stat_StatFeralAttackPower
							}
							if stat == proto.Stat_StatArmorPenetration || stat == proto.Stat_StatSpellPiercing {
								// Make these not negative
								value = math.Abs(value)
							}
						}
						outStats[stat] += value
					}
				}
			}
		case ITEM_ENCHANTMENT_COMBAT_SPELL:
			// Not processed (chance on hit, ignore for now)
		case ITEM_ENCHANTMENT_USE_SPELL:
			// Not processed
		}
	}
}

const rangedWeaponSubclassMask = rangedMask | ITEM_SUBCLASS_BIT_WEAPON_THROWN | ITEM_SUBCLASS_BIT_WEAPON_WAND

// Adds the percents an equip spell states for hit, crit, block, dodge, parry and attack and cast speed
// to pseudoStats, which is indexed by proto.PseudoStat, and reports whether it added any. Only auras on
// the wearer count: Atiesh's 28142 is a party aura and reaches the sim as a raid buff.
//
// A spell that names weapons applies its hit and weapon crit only to attacks with them. Melee is
// credited when it names no weapon or a melee one and ranged when it names no weapon or a ranged one,
// and ranged is written as the total the character sheet shows, melee share included, which
// stats.FromPseudoStatsProto takes back out. So 22780 (Biznicks 247x128 Accurascope, hit restricted to
// bows, guns and crossbows) is ranged hit alone, and 1310308 (SAF-T Ultra Precision Scope, the all-crit
// aura under the same restriction) is ranged crit alone, with no spell crit.
func AddEquipSpellPseudoStats(pseudoStats []float64, spellID int) bool {
	spell := dbcInstance.Spells[spellID]
	namesWeapons := spell.EquippedItemClass == ITEM_CLASS_WEAPON && spell.EquippedItemSubclass != 0
	melee := !namesWeapons || spell.EquippedItemSubclass&^rangedWeaponSubclassMask != 0
	ranged := !namesWeapons || spell.EquippedItemSubclass&rangedWeaponSubclassMask != 0

	added := false
	add := func(pseudoStat proto.PseudoStat, value float64) {
		pseudoStats[pseudoStat] += value
		added = true
	}
	addWeapon := func(meleeStat, rangedStat proto.PseudoStat, value float64) {
		if melee {
			add(meleeStat, value)
		}
		if ranged {
			add(rangedStat, value)
		}
	}

	for _, effect := range dbcInstance.SpellEffectsInOrder(spellID) {
		if effect.EffectType != dbcenums.E_APPLY_AURA {
			continue
		}
		value := effect.EffectBasePoints + effect.EffectDieSides

		switch effect.EffectAura {
		case dbcenums.A_MOD_HIT_CHANCE:
			addWeapon(proto.PseudoStat_PseudoStatMeleeHitPercent, proto.PseudoStat_PseudoStatRangedHitPercent, value)
		case dbcenums.A_MOD_WEAPON_CRIT_PERCENT:
			addWeapon(proto.PseudoStat_PseudoStatMeleeCritPercent, proto.PseudoStat_PseudoStatRangedCritPercent, value)
		case dbcenums.A_MOD_CRIT_PCT:
			addWeapon(proto.PseudoStat_PseudoStatMeleeCritPercent, proto.PseudoStat_PseudoStatRangedCritPercent, value)
			if !namesWeapons {
				add(proto.PseudoStat_PseudoStatSpellCritPercent, value)
			}
		case dbcenums.A_MOD_SPELL_HIT_CHANCE:
			add(proto.PseudoStat_PseudoStatSpellHitPercent, value)
		case dbcenums.A_MOD_SPELL_CRIT_CHANCE:
			add(proto.PseudoStat_PseudoStatSpellCritPercent, value)
		case dbcenums.A_MOD_BLOCK_PERCENT:
			add(proto.PseudoStat_PseudoStatBlockPercent, value)
		case dbcenums.A_MOD_DODGE_PERCENT:
			add(proto.PseudoStat_PseudoStatDodgePercent, value)
		case dbcenums.A_MOD_PARRY_PERCENT:
			add(proto.PseudoStat_PseudoStatParryPercent, value)
		}

		// 1293881 (Pendulum of Doom) states a zero melee haste and changes the speed through its procs.
		if value != 0 {
			for _, pseudoStat := range equipSpellHasteAuras[effect.EffectAura] {
				add(pseudoStat, value)
			}
		}
	}

	return added
}

// Unlike ranged hit and crit, ranged haste is not a total including melee: the sim applies melee,
// ranged and cast speed each on its own.
var equipSpellHasteAuras = map[EffectAuraType][]proto.PseudoStat{
	dbcenums.A_MOD_MELEE_HASTE_3:           {proto.PseudoStat_PseudoStatMeleeHastePercent},
	dbcenums.A_MOD_RANGED_HASTE:            {proto.PseudoStat_PseudoStatRangedHastePercent},
	dbcenums.A_MOD_MELEE_RANGED_HASTE_2:    {proto.PseudoStat_PseudoStatMeleeHastePercent, proto.PseudoStat_PseudoStatRangedHastePercent},
	dbcenums.A_MOD_CASTING_SPEED_NOT_STACK: {proto.PseudoStat_PseudoStatSpellHastePercent},
}

const skillLineDefense = 95

// Adds the defense skill and flat block value an equip spell states to outStats, and reports whether it
// added any. 24148 (Presence of Might) states A_MOD_SKILL 7 on the Defense skill line and
// A_MOD_BLOCK_VALUE_FLAT 15.
func AddEquipSpellStats(outStats *stats.Stats, spellID int) bool {
	added := false
	for _, effect := range dbcInstance.SpellEffectsInOrder(spellID) {
		if effect.EffectType != dbcenums.E_APPLY_AURA {
			continue
		}
		value := effect.EffectBasePoints + effect.EffectDieSides

		switch {
		case effect.EffectAura == dbcenums.A_MOD_SKILL && effect.EffectMiscValues[0] == skillLineDefense:
			outStats[proto.Stat_StatDefenseRating] += value * core.DefenseRatingPerDefenseLevel
		case effect.EffectAura == dbcenums.A_MOD_BLOCK_VALUE_FLAT:
			outStats[proto.Stat_StatBlockValue] += value
		default:
			continue
		}
		added = true
	}

	return added
}

func ConvertEffectAuraToStatIndex(effectAura EffectAuraType, effectMisc int) proto.Stat {
	switch effectAura {
	case dbcenums.A_MOD_ATTACK_POWER:
		return proto.Stat_StatAttackPower
	case dbcenums.A_MOD_RANGED_ATTACK_POWER:
		return proto.Stat_StatRangedAttackPower
	case dbcenums.A_MOD_DAMAGE_DONE:
		return ConvertSpellDamageFlagToSchoolDamageStat(effectMisc)
	case dbcenums.A_MOD_HEALING_DONE:
		return proto.Stat_StatHealingPower
	case dbcenums.A_MOD_INCREASE_HEALTH:
		return proto.Stat_StatHealth
	case dbcenums.A_MOD_TARGET_RESISTANCE:
		return ConvertTargetResistanceFlagToPenetrationStat(effectMisc)
	case dbcenums.A_MOD_RESISTANCE:
		return ConvertResistanceFlagToResistanceStat(effectMisc)
	case dbcenums.A_MOD_RATING:
		return ConvertModRatingFlagToRatingStat(effectMisc)
	case dbcenums.A_MOD_SHIELD_BLOCKVALUE:
		return proto.Stat_StatBlockValue
	case dbcenums.A_MOD_POWER_REGEN:
		return proto.Stat_StatMP5
	default:
		return -1
	}
}

func ConvertResistanceFlagToResistanceStat(flag int) proto.Stat {
	school := SpellSchool(flag)
	if school == ALL_SPELL_DAMAGE || school == SPELL_PENETRATION {
		// All 5 Magic School resist; return -2 to be handled as special case
		// 124 excludes "Holy" which isn't a resist anyways
		return -2
	}
	for schoolType, stat := range SpellSchoolToResistanceStat {
		if school.Has(schoolType) {
			return stat
		}
	}
	return -1
}

func ConvertTargetResistanceFlagToPenetrationStat(flag int) proto.Stat {
	switch flag {
	case 1:
		return proto.Stat_StatArmorPenetration
	default:
		return proto.Stat_StatSpellPiercing
	}
}

func ConvertSpellDamageFlagToSchoolDamageStat(flag int) proto.Stat {
	switch flag {
	case 1:
		return proto.Stat_StatPhysicalDamage
	case 2:
		return proto.Stat_StatHolyDamage
	case 4:
		return proto.Stat_StatFireDamage
	case 8:
		return proto.Stat_StatNatureDamage
	case 16:
		return proto.Stat_StatFrostDamage
	case 32:
		return proto.Stat_StatShadowDamage
	case 64:
		return proto.Stat_StatArcaneDamage
	default:
		return proto.Stat_StatSpellDamage
	}
}

func ConvertModRatingFlagToRatingStat(flag int) proto.Stat {
	switch flag {
	case 2:
		return proto.Stat_StatDefenseRating
	case 4:
		return proto.Stat_StatDodgeRating
	case 8:
		return proto.Stat_StatParryRating
	case 16:
		return proto.Stat_StatBlockRating
	case 64:
		// The forbidden "Only Ranged Hit". There's a single instance of this (Enchant 2523, SpellID 22780).
		return -1
	case 96:
		return proto.Stat_StatMeleeHitRating
	case 128:
		return proto.Stat_StatSpellHitRating
	case 256:
		return proto.Stat_StatFireResistance
	case 512:
		// The forbidden "Only Ranged Crit". Only two of these exist, one is low level gloves another is an enchant.
		return -1
	case 768:
		return proto.Stat_StatMeleeCritRating
	case 1024:
		return proto.Stat_StatSpellCritRating
	case 49152:
		return proto.Stat_StatResilienceRating
	case 131072:
		return proto.Stat_StatMeleeHasteRating
	case 393216:
		return proto.Stat_StatMeleeHasteRating
	case 524288:
		return proto.Stat_StatMeleeHasteRating
	default:
		println("UNHANDLED RATING FLAG: " + strconv.Itoa(flag))
		return -1
	}
}
