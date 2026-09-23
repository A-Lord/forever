package core

import (
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func registerStaticImbue(agent Agent, imbueId int32) {
	character := agent.GetCharacter()
	switch imbueId {
	case 25123: // Mana Oil
		character.AddStat(stats.HealingPower, 30)
		character.AddStat(stats.MP5, 15)
	case 25122: // Briliant Wizard Oil
		character.AddStat(stats.SpellDamage, 36)
		character.AddStat(stats.HealingPower, 36)
		character.AddStat(stats.SpellCritPercent, 1)
	case 29453, 34340: // Adamantite Sharpening Stone / Adamantite Weightstone
		character.AddStat(stats.MeleeCritRating, 14)
		for _, weapon := range []*Weapon{character.AutoAttacks.MH(), character.AutoAttacks.OH(), character.AutoAttacks.Ranged()} {
			if weapon != nil {
				weapon.BaseDamageMin += 12
				weapon.BaseDamageMax += 12
			}
		}
		// Keep Ranged Crit the same
		character.AddStat(stats.RangedCritPercent, -(14 / PhysicalCritRatingPerCritPercent))
	case 28891: // Consecrated Sharpening Stone
		character.Env.RegisterPostFinalizeEffect(func() {
			for _, at := range character.AttackTables {
				at.MobTypeBonusStats[proto.MobType_MobTypeUndead] = at.MobTypeBonusStats[proto.MobType_MobTypeUndead].Add(stats.Stats{
					stats.AttackPower:       100,
					stats.RangedAttackPower: 100,
				})
			}
		})
	}
}
