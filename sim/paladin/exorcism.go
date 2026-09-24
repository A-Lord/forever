package paladin

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

var ExorcismRankMap = spellData.Exorcism

// Exorcism
// https://www.wowhead.com/forever/spell=10314
//
// Causes 502 Holy damage to an Undead or Demon target.
//
// spellData.Exorcism holds the six trainer ranks, 879 to 10314. The client carries a second ladder,
// 415068 to 415073, that the Season of Discovery passive Exorcist (415076) swaps onto the action
// bar so the spell can hit any target; nothing in Forever teaches Exorcist.
func (paladin *Paladin) registerExorcism(row shared.SpellData) {
	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: row.SpellID},
		SpellSchool:    row.SpellSchool,
		DefenseType:    row.DefenseType,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskExorcism,
		Rank:           row.Rank,
		MaxRange:       row.MaxRange,

		ManaCost: manaCost(row),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: row.GCD,
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.exorcismTimer),
				Duration: row.Cooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: row.Direct.BonusCoefficient(),

		ExtraCastCondition: func(_ *core.Simulation, target *core.Unit) bool {
			return target.MobType == proto.MobType_MobTypeUndead || target.MobType == proto.MobType_MobTypeDemon
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, row.Direct.Damage(sim), spell.OutcomeMagicHitAndCrit)
		},
	})
}
