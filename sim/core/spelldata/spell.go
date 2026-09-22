package spelldata

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// SpellPower.PowerType 1. The client states rage on a 0-1000 bar, so a 15 rage cost reads 150 here;
// mana, focus and energy are stated in whole points - see NormalizePowerCost in tools/database.
const powerTypeRage int8 = 1

// What Power answers for a bar the spell does not use. Shared, like Nil and NilEffect, so a caller
// must not write through it.
var nilPower = &Power{}

func (s *Spell) CastTime() time.Duration {
	return millis(s.CastTimeMs)
}

// SpellCooldowns.RecoveryTime. A spell gated by its category instead - Fire Blast and Cone of Cold
// share one - states that in CategoryCooldown.
func (s *Spell) Cooldown() time.Duration {
	return millis(s.CooldownMs)
}

func (s *Spell) CategoryCooldown() time.Duration {
	return millis(s.CategoryCooldownMs)
}

func (s *Spell) GCD() time.Duration {
	return millis(s.GCDMs)
}

// The client states a permanent aura as -1, which is core.NeverExpires here.
func (s *Spell) Duration() time.Duration {
	if s.DurationMs == -1 {
		return core.NeverExpires
	}
	return millis(s.DurationMs)
}

// SpellAuraOptions.ProcCategoryRecovery: the internal cooldown between two procs.
func (s *Spell) ICD() time.Duration {
	return millis(s.ICDMs)
}

// core.SpellSchool holds the client's own bit values - see the comment on SpellSchool in
// sim/core/flags.go - so the mask converts without a translation table.
func (s *Spell) SpellSchool() core.SpellSchool {
	return core.SpellSchool(s.School)
}

// SpellCategories.DefenseType counts none, magic, melee, ranged in that order, which is the order
// core.DefenseType is declared in.
func (s *Spell) DefenseTypeCore() core.DefenseType {
	return core.DefenseType(s.DefenseType)
}

// The i-th effect the row carries, counted from 1 by position rather than by the client's
// EffectIndex, which has gaps. Out of range answers NilEffect.
func (s *Spell) EffectN(i int) *Effect {
	if i < 1 || i > len(s.Effects) {
		return NilEffect
	}
	return &s.Effects[i-1]
}

// The one effect with this aura and misc value. Panics when none matches, and when two do: reading
// the first of several silently is the bug this shape exists to prevent, and EffectN is the way past
// it.
func (s *Spell) Effect(aura dbcenums.EffectAuraType, misc int32) *Effect {
	found, matches := -1, 0
	for i := range s.Effects {
		if s.Effects[i].Aura == aura && s.Effects[i].Misc == misc {
			matches++
			if found < 0 {
				found = i
			}
		}
	}
	if matches > 1 {
		panic(fmt.Sprintf("spell %d has %d effects with aura %d misc %d - index them instead",
			s.ID, matches, aura, misc))
	}
	if found < 0 {
		panic(fmt.Sprintf("spell %d has no effect with aura %d misc %d, in %d effects",
			s.ID, aura, misc, len(s.Effects)))
	}
	return &s.Effects[found]
}

// The first effect matching all three, or NilEffect. Zero matches anything it is given as: an aura of
// 0 on a direct effect is what the client states there.
func (s *Spell) FindEffect(typ dbcenums.SpellEffectType, aura dbcenums.EffectAuraType, misc int32) *Effect {
	for i := range s.Effects {
		e := &s.Effects[i]
		if e.Type == typ && e.Aura == aura && e.Misc == misc {
			return e
		}
	}
	return NilEffect
}

// The effect carrying the spell's direct damage, whether it states an amount or a weapon multiplier.
// It sits on any of the weapon effects as readily as on school damage: a weapon effect states a
// multiplier or a flat bonus rather than an amount.
func (s *Spell) DamageEffect() *Effect {
	return s.firstOfType(dbcenums.E_SCHOOL_DAMAGE, dbcenums.E_WEAPON_DAMAGE,
		dbcenums.E_WEAPON_PERCENT_DAMAGE, dbcenums.E_NORMALIZED_WEAPON_DMG,
		dbcenums.E_WEAPON_DAMAGE_NOSCHOOL)
}

func (s *Spell) HealEffect() *Effect {
	return s.firstOfType(dbcenums.E_HEAL)
}

func (s *Spell) EnergizeEffect() *Effect {
	return s.firstOfType(dbcenums.E_ENERGIZE)
}

// The effect that ticks: an aura application whose aura carries a per-tick value, which is damage,
// healing, mana or a spell fired each tick.
func (s *Spell) PeriodicEffect() *Effect {
	for i := range s.Effects {
		e := &s.Effects[i]
		if e.Type != dbcenums.E_APPLY_AURA {
			continue
		}
		switch e.Aura {
		case dbcenums.A_PERIODIC_DAMAGE, dbcenums.A_PERIODIC_HEAL, dbcenums.A_PERIODIC_ENERGIZE,
			dbcenums.A_PERIODIC_TRIGGER_SPELL, dbcenums.A_PERIODIC_LEECH:
			return e
		}
	}
	return NilEffect
}

func (s *Spell) firstOfType(types ...dbcenums.SpellEffectType) *Effect {
	for i := range s.Effects {
		for _, t := range types {
			if s.Effects[i].Type == t {
				return &s.Effects[i]
			}
		}
	}
	return NilEffect
}

// The cost out of this bar, or a zero Power where the spell does not use it. Both the row's own and
// the shared zero one are the store's, so a caller must not write through what it gets back.
func (s *Spell) Power(t int8) *Power {
	for i := range s.Powers {
		if s.Powers[i].Type == t {
			return &s.Powers[i]
		}
	}
	return nilPower
}

// The cost in the units the sim spends: rage off the client's 0-1000 bar, everything else as stated.
func (s *Spell) PowerCost(t int8) float64 {
	cost := float64(s.Power(t).Cost)
	if t == powerTypeRage {
		return cost / 10
	}
	return cost
}

// The cost of the bar the spell spends - the first SpellPower row - in the units PowerCost converts
// to. 0 for a spell with no SpellPower row.
func (s *Spell) Cost() float64 {
	if len(s.Powers) == 0 {
		return 0
	}
	return s.PowerCost(s.Powers[0].Type)
}

// Spell.NameSubtext_lang's "Rank N" as a number; 0 for a spell the client shows no rank on, whose
// subtext is "Passive" or empty.
func (s *Spell) RankNumber() int32 {
	return rankOf(s)
}

func (s *Spell) HasLabel(id int16) bool {
	return slices.Contains(s.Labels, id)
}

// Whether the modifier effect names this spell. A label-keyed modifier aura names its spells through
// SpellLabel instead of through a class mask; matching those means keying on the aura, since the label
// sits in the effect's misc value only for that family of auras.
// TODO: match e.Misc against s.Labels for the auras that name their targets by label, the way SimC's
// affected_by_label does, once the sim registers a talent that uses one: A_MOD_RECOVERY_RATE_BY_SPELL_LABEL
// 143, A_SUPPRESS_ITEM_PASSIVE_EFFECT_BY_SPELL_LABEL 182, A_ADD_PCT_MODIFIER_BY_SPELL_LABEL 218,
// A_ADD_FLAT_MODIFIER_BY_SPELL_LABEL 219, A_CAST_WHILE_WALKING_BY_SPELL_LABEL 307,
// A_MOD_AURA_TIME_RATE_BY_SPELL_LABEL 470 and A_MOD_DAMAGE_TAKEN_FROM_CASTER_BY_LABEL 507.
func (s *Spell) AffectedBy(e *Effect) bool {
	return s.ClassFlags.Matches(e.ClassFlags)
}

// The spells the tooltip names, in the order it names them. An id the store does not carry is left
// out rather than answered as Nil.
func (s *Spell) Refs() []*Spell {
	return resolve(s.RefIDs)
}

// The spells whose effects fire this one, from the trigger index.
func (s *Spell) Drivers() []*Spell {
	return resolve(drivers[s.ID])
}

// Every spell this one's effects fire, deduped, in effect order.
func (s *Spell) Triggered() []*Spell {
	var ids []int32
	for _, e := range s.Effects {
		if e.TriggerID == 0 {
			continue
		}
		if !slices.Contains(ids, e.TriggerID) {
			ids = append(ids, e.TriggerID)
		}
	}
	return resolve(ids)
}

// The outcome a periodic tick rolls: a tick that can crit where the client marks Periodic Can Crit, a
// plain tick otherwise, on the hit table the row's defense type names.
func (s *Spell) TickOutcome(dot *core.Dot) core.OutcomeApplier {
	switch tickOutcomeKind(s.PeriodicCanCrit(), s.DefenseTypeCore() == core.DefenseTypeMagic) {
	case tickOutcomeMagicCrit:
		return dot.Spell.OutcomeTickMagicHitAndCrit
	case tickOutcomePhysicalCrit:
		return dot.Spell.OutcomeTickPhysicalCrit
	case tickOutcomeMagicHit:
		return dot.OutcomeTickMagicHit
	default:
		return dot.OutcomeTick
	}
}

// Which of the four ticks the two flags pick. Split out so the choice can be asserted without a Dot.
const (
	tickOutcomeMagicCrit = iota
	tickOutcomePhysicalCrit
	tickOutcomeMagicHit
	tickOutcomePlain
)

func tickOutcomeKind(canCrit bool, magic bool) int {
	switch {
	case canCrit && magic:
		return tickOutcomeMagicCrit
	case canCrit:
		return tickOutcomePhysicalCrit
	case magic:
		return tickOutcomeMagicHit
	default:
		return tickOutcomePlain
	}
}

func resolve(ids []int32) []*Spell {
	var out []*Spell
	for _, id := range ids {
		if s := Find(id); s != Nil {
			out = append(out, s)
		}
	}
	return out
}

func millis(ms int32) time.Duration {
	return time.Duration(ms) * time.Millisecond
}
