package shared

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The auras granting a stat against the creature types their misc value masks, which the parser
// does not attach.
var versusAuras = map[dbcenums.EffectAuraType]stats.Stat{
	dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS:  stats.AttackPower,
	dbcenums.A_MOD_RANGED_ATTACK_POWER_VERSUS: stats.RangedAttackPower,
	dbcenums.A_MOD_FLAT_SPELL_DAMAGE_VERSUS:   stats.SpellDamage,
}

// The mob type of each bit of a versus aura's creature type mask.
var versusMobTypes = map[int32]proto.MobType{
	1 << 0: proto.MobType_MobTypeBeast,
	1 << 1: proto.MobType_MobTypeDragonkin,
	1 << 2: proto.MobType_MobTypeDemon,
	1 << 3: proto.MobType_MobTypeElemental,
	1 << 4: proto.MobType_MobTypeGiant,
	1 << 5: proto.MobType_MobTypeUndead,
	1 << 6: proto.MobType_MobTypeHumanoid,
	1 << 8: proto.MobType_MobTypeMechanical,
}

// A potion whose aura grants attack power or spell damage against creature types, as a
// hand-written effect: Potion of Demonslaying 9224.
func NewSpellDataVersusPotion(itemID int32) {
	auras := make([]dbcenums.EffectAuraType, 0, len(versusAuras))
	for aura := range versusAuras {
		auras = append(auras, aura)
	}
	registerSpellDataConsumable(itemID, core.NewConsumableEffect, attachedAuras{
		auras: auras,
		attach: func(character *core.Character, aura *core.Aura, row *spelldata.Spell) {
			attachVersusBonus(character, aura, versusBonus(row, character.Level))
		},
	})
}

// A temporary enchant whose equip spell grants attack power or spell damage against creature types,
// as a hand-written effect: Consecrated Sharpening Stone 23122, whose enchant 2684 casts 28893, and
// Blessed Wizard Oil 23123, whose enchant 2685 casts 28890. Each hand core put the enchant on carries
// its own copy of the equip aura, so the bonus counts once per hand.
func NewSpellDataVersusTemporaryEnchant(itemID, equipSpellID int32) {
	core.NewConsumableEffect(itemID, func(agent core.Agent) {
		character := agent.GetCharacter()
		hands := applyTemporaryEnchant(character, core.GetConsumableByID(itemID))
		if hands == 0 {
			return
		}

		row := spelldata.MustFind(equipSpellID)
		bonus := versusBonus(row, character.Level)
		for mobType, granted := range bonus {
			bonus[mobType] = granted.Multiply(float64(hands))
		}
		aura := character.GetOrRegisterAura(spelldata.AuraConfig(row, spelldata.Permanent()))
		attachVersusBonus(character, aura, bonus)
	})
}

// The stats the row's versus auras grant, whatever creature types they name.
func VersusStats(row *spelldata.Spell, level int32) stats.Stats {
	var granted stats.Stats
	for i := range row.Effects {
		e := &row.Effects[i]
		if stat, ok := versusAuras[e.Aura]; ok && e.Type == dbcenums.E_APPLY_AURA {
			granted[stat] += e.Average(level)
		}
	}
	return granted
}

func versusBonus(row *spelldata.Spell, level int32) map[proto.MobType]stats.Stats {
	bonus := map[proto.MobType]stats.Stats{}
	for i := range row.Effects {
		e := &row.Effects[i]
		stat, ok := versusAuras[e.Aura]
		if e.Type != dbcenums.E_APPLY_AURA || !ok {
			continue
		}
		var granted stats.Stats
		granted[stat] = e.Average(level)
		for bit, mobType := range versusMobTypes {
			if e.Misc&bit != 0 {
				bonus[mobType] = bonus[mobType].Add(granted)
			}
		}
	}
	return bonus
}

func attachVersusBonus(character *core.Character, aura *core.Aura, bonus map[proto.MobType]stats.Stats) {
	aura.ApplyOnGain(func(_ *core.Aura, _ *core.Simulation) {
		for _, at := range character.AttackTables {
			for mobType, bonusStats := range bonus {
				at.MobTypeBonusStats[mobType] = at.MobTypeBonusStats[mobType].Add(bonusStats)
			}
		}
	}).ApplyOnExpire(func(_ *core.Aura, _ *core.Simulation) {
		for _, at := range character.AttackTables {
			for mobType, bonusStats := range bonus {
				at.MobTypeBonusStats[mobType] = at.MobTypeBonusStats[mobType].Subtract(bonusStats)
			}
		}
	})
}
