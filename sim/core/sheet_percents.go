package core

import (
	"github.com/wowsims/forever/sim/core/stats"
)

type sheetAttackStats struct {
	HitPercent, CritPercent, HitRating, CritRating float64
}

// The hit and crit the character sheet shows for the auto attack with this proc mask: the physical
// percent, the ranged percent for a ranged probe, and the hit or crit mods that reach the attack. The
// dual-wield miss penalty is not in them.
//
// The rating is the one hit or crit rating, plus the ranged percent and the mods converted from a
// rating, each at the rating per percent. A mod stated as a percent is in the percent alone.
func (character *Character) sheetAttackStats(probe ProcMask) sheetAttackStats {
	var channelHit, channelCrit float64
	if probe.Matches(ProcMaskRanged) {
		channelHit, channelCrit = character.GetStat(stats.RangedHitPercent), character.GetStat(stats.RangedCritPercent)
	}
	scoped := character.scopedPercents(probe)
	return sheetAttackStats{
		HitPercent:  character.GetStat(stats.PhysicalHitPercent) + channelHit + scoped.hit,
		CritPercent: character.GetStat(stats.PhysicalCritPercent) + channelCrit + scoped.crit,
		HitRating:   character.GetStat(stats.HitRating) + (channelHit+scoped.hitFromRating)*HitRatingPerHitPercent,
		CritRating:  character.GetStat(stats.CritRating) + (channelCrit+scoped.critFromRating)*CritRatingPerCritPercent,
	}
}
