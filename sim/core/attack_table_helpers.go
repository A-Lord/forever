package core

import (
	"slices"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// The attack tables exist once the environment is set up, so the Unit helpers act on the tables of
// that moment: a caller runs them from a post-finalize effect or an aura's gain and expire.

// Adds stats the unit has against every defender of the mob types; a negative bonus takes them away.
func (unit *Unit) AddMobTypeBonusStats(bonus stats.Stats, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		for _, mobType := range mobTypes {
			at.MobTypeBonusStats[mobType] = at.MobTypeBonusStats[mobType].Add(bonus)
		}
	}
}

// Multiplies the damage the unit deals to every defender of the mob types, and the crit damage too
// where crit is set.
func (unit *Unit) MultiplyDamageVersus(mult float64, crit bool, mobTypes ...proto.MobType) {
	for _, at := range unit.AttackTables {
		if slices.Contains(mobTypes, at.Defender.MobType) {
			at.DamageDealtMultiplier *= mult
			if crit {
				at.CritMultiplier *= mult
			}
		}
	}
}

func (unit *Unit) divideDamageVersus(mult float64, crit bool, mobTypes []proto.MobType) {
	for _, at := range unit.AttackTables {
		if slices.Contains(mobTypes, at.Defender.MobType) {
			at.DamageDealtMultiplier /= mult
			if crit {
				at.CritMultiplier /= mult
			}
		}
	}
}

// Adds to the share of every defender's armor the unit's attacks ignore; a negative delta takes it away.
func (unit *Unit) AddArmorIgnoreFactor(delta float64) {
	for _, at := range unit.AttackTables {
		at.ArmorIgnoreFactor += delta
	}
}

// The stats against the mob types while the aura is up.
// Returns parent aura for chaining
func (parentAura *Aura) AttachMobTypeBonusStats(bonus stats.Stats, mobTypes ...proto.MobType) *Aura {
	parentAura.ApplyOnGain(func(aura *Aura, _ *Simulation) {
		aura.Unit.AddMobTypeBonusStats(bonus, mobTypes...)
	})

	parentAura.ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		aura.Unit.AddMobTypeBonusStats(bonus.Invert(), mobTypes...)
	})

	if parentAura.IsActive() {
		parentAura.Unit.AddMobTypeBonusStats(bonus, mobTypes...)
	}

	return parentAura
}

// The damage multiplier against the mob types while the aura is up.
// Returns parent aura for chaining
func (parentAura *Aura) AttachDamageMultiplierVersus(mult float64, crit bool, mobTypes ...proto.MobType) *Aura {
	parentAura.ApplyOnGain(func(aura *Aura, _ *Simulation) {
		aura.Unit.MultiplyDamageVersus(mult, crit, mobTypes...)
	})

	parentAura.ApplyOnExpire(func(aura *Aura, _ *Simulation) {
		aura.Unit.divideDamageVersus(mult, crit, mobTypes)
	})

	if parentAura.IsActive() {
		parentAura.Unit.MultiplyDamageVersus(mult, crit, mobTypes...)
	}

	return parentAura
}
