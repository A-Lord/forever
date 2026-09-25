package core

import (
	"github.com/wowsims/forever/sim/core/stats"
)

type sheetAttackStats struct {
	HitPercent, CritPercent, HitRating, CritRating float64
}

// The hit and crit the character sheet shows for the auto attack with this proc mask: the physical
// percent, the channel's own percent, and the hit or crit mods that reach the attack. The dual-wield
// miss penalty is not in them.
//
// The rating is the one hit or crit rating, plus the channel's own percent and the mods converted
// from a rating, each at the rating per percent. A mod stated as a percent is in the percent alone.
func (character *Character) sheetAttackStats(probe ProcMask, channelHit, channelCrit float64) sheetAttackStats {
	return sheetAttackStats{
		HitPercent: character.GetStat(stats.PhysicalHitPercent) + channelHit +
			character.ScopedBonusPercent(SpellMod_BonusHit_Percent, probe),
		CritPercent: character.GetStat(stats.PhysicalCritPercent) + channelCrit +
			character.ScopedBonusPercent(SpellMod_BonusCrit_Percent, probe),
		HitRating: character.GetStat(stats.HitRating) +
			(channelHit+character.ScopedRatingPercent(SpellMod_BonusHit_Percent, probe))*HitRatingPerHitPercent,
		CritRating: character.GetStat(stats.CritRating) +
			(channelCrit+character.ScopedRatingPercent(SpellMod_BonusCrit_Percent, probe))*CritRatingPerCritPercent,
	}
}

func (character *Character) sheetMainHandStats() sheetAttackStats {
	return character.sheetAttackStats(ProcMaskMeleeMHAuto, 0, 0)
}

func (character *Character) sheetOffHandStats() sheetAttackStats {
	if !character.AutoAttacks.IsDualWielding {
		return sheetAttackStats{}
	}
	return character.sheetAttackStats(ProcMaskMeleeOHAuto, 0, 0)
}

func (character *Character) sheetRangedStats() sheetAttackStats {
	return character.sheetAttackStats(ProcMaskRangedAuto, character.GetStat(stats.RangedHitPercent), character.GetStat(stats.RangedCritPercent))
}
