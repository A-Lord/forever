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
// The rating is the one hit or crit rating plus the rest of the percent at the rating per percent: a
// scoped mod keeps no record of whether it came from a rating or a percent, so each counts as the
// rating it is worth.
func (character *Character) sheetAttackStats(probe ProcMask, channelHit, channelCrit float64) sheetAttackStats {
	hit := channelHit + character.ScopedBonusPercent(SpellMod_BonusHit_Percent, probe)
	crit := channelCrit + character.ScopedBonusPercent(SpellMod_BonusCrit_Percent, probe)
	return sheetAttackStats{
		HitPercent:  character.GetStat(stats.PhysicalHitPercent) + hit,
		CritPercent: character.GetStat(stats.PhysicalCritPercent) + crit,
		HitRating:   character.GetStat(stats.HitRating) + hit*HitRatingPerHitPercent,
		CritRating:  character.GetStat(stats.CritRating) + crit*CritRatingPerCritPercent,
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
