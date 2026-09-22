package spelldata

import (
	"fmt"
	"os"
	"slices"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// How a parse is narrowed: which effects it reads, what gates them and whether their values follow
// the aura's stacks.
type ParseOpt func(*parseOptions)

type parseOptions struct {
	only         []int32
	skip         []int32
	cond         func() bool
	ignoreStacks bool
}

// Only these effects, counted from 1 by position the way EffectN counts.
func Effects(idx ...int32) ParseOpt {
	return func(o *parseOptions) {
		o.only = append(o.only, idx...)
	}
}

func SkipEffects(idx ...int32) ParseOpt {
	return func(o *parseOptions) {
		o.skip = append(o.skip, idx...)
	}
}

// A condition every attachment is gated on, beyond the aura being up: it is read when the aura is
// gained and whenever the caller calls Refresh, and nothing is active while it answers false.
func Conditional(fn func() bool) ParseOpt {
	return func(o *parseOptions) {
		o.cond = fn
	}
}

// The values do not follow the aura's stacks, which is what a row that states charges rather than
// cumulative stacks means.
func IgnoreStacks() ParseOpt {
	return func(o *parseOptions) {
		o.ignoreStacks = true
	}
}

// One attachment the parse made, for a log or a test: the effect it came from, the sim kind it
// became and the value it carries in the sim's own units. A time value is stated in milliseconds.
type Applied struct {
	Effect *Effect
	Kind   string
	Value  float64
}

// What one parse did. Skipped holds the aura effects the table does not know, which are the ones a
// port still has to wire by hand.
type Parsed struct {
	Applied []Applied
	Skipped []*Effect

	attachments []*attachment
	aura        *core.Aura
	cond        func() bool
	stacking    bool
}

// Everything the row's aura effects say, attached to an aura: the mods, stat buffs and pseudo-stat
// multipliers follow the aura's gain and expiry, and their values follow its stacks where the row
// states cumulative ones.
//
// The effects act on the aura's own unit, and the character is that unit's, for the few rows whose
// helper lives on a character rather than on a unit. An aura on a unit with no character of its own -
// a debuff on an enemy - takes a nil character and has those rows skipped and reported.
//
// An aura that is already up when it is parsed is caught up the way core's Attach helpers catch one
// up, except for the rows that need a Simulation to act: the haste multipliers stay off until the
// aura is applied again.
func ParseEffects(character *core.Character, aura *core.Aura, s *Spell, opts ...ParseOpt) *Parsed {
	if aura == nil {
		return &Parsed{}
	}
	return parse(aura.Unit, character, aura, s, opts)
}

// Everything the row's aura effects say, applied to the character now: a talent or a passive, which
// the client states as an aura the unit is never without.
func ParseStatic(character *core.Character, s *Spell, opts ...ParseOpt) *Parsed {
	if character == nil {
		return &Parsed{}
	}
	return parse(&character.Unit, character, nil, s, opts)
}

func parse(unit *core.Unit, character *core.Character, aura *core.Aura, s *Spell, opts []ParseOpt) *Parsed {
	parsed := &Parsed{}
	if unit == nil || s == nil || s == Nil || len(s.Effects) == 0 {
		return parsed
	}

	o := &parseOptions{}
	for _, opt := range opts {
		opt(o)
	}

	p := &parser{
		unit:        unit,
		character:   character,
		spell:       s,
		static:      aura == nil,
		conditional: o.cond != nil,
		stacking:    s.MaxStack > 0 && !o.ignoreStacks,
	}

	folded := foldedDotEffects(s, o)

	// The amount is the caster's, so a debuff on an enemy is still priced at the character's level
	// and not at the level of the unit it sits on. Without a character there is only the unit.
	level := unit.Level
	if character != nil {
		level = character.Level
	}

	for i := range s.Effects {
		e := &s.Effects[i]
		if !o.reads(int32(i+1)) || !appliesAura(e.Type) {
			continue
		}

		value := e.Average(level)

		if slices.Contains(folded, e) {
			parsed.Applied = append(parsed.Applied, Applied{Effect: e,
				Kind: "folded-into SpellMod_DamageDone_Flat", Value: value / 100})
			parsed.attachments = append(parsed.attachments, nil)
			continue
		}

		attached := auraTable[e.Aura]
		if attached == nil {
			parsed.skip(s, i, e)
			continue
		}

		a := attached(p, e, value)
		if a == nil {
			parsed.skip(s, i, e)
			continue
		}

		parsed.Applied = append(parsed.Applied, Applied{Effect: e, Kind: a.kind, Value: a.value})
		parsed.attachments = append(parsed.attachments, a)
	}

	parsed.aura = aura
	parsed.cond = o.cond
	parsed.stacking = p.stacking

	if aura == nil {
		parsed.set(nil, parsed.level())
		return parsed
	}

	aura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		parsed.set(sim, parsed.level())
	}).ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		parsed.set(sim, 0)
	})

	if p.stacking {
		aura.ApplyOnStacksChange(func(_ *core.Aura, sim *core.Simulation, _ int32, _ int32) {
			parsed.set(sim, parsed.level())
		})
	}

	if aura.IsActive() {
		parsed.catchUp(parsed.level())
	}

	return parsed
}

// Reads the condition again and turns the attachments on or off accordingly, for a caller whose
// condition changed under a parse that is already in place.
func (p *Parsed) Refresh(sim *core.Simulation) {
	p.set(sim, p.level())
}

// What the attachments are worth right now: nothing while the condition answers false or the aura is
// down, the plain value where the row does not stack, and once per stack where it does.
func (p *Parsed) level() float64 {
	if p.cond != nil && !p.cond() {
		return 0
	}
	if p.aura != nil && !p.aura.IsActive() {
		return 0
	}
	if p.stacking && p.aura != nil {
		if stacks := p.aura.GetStacks(); stacks > 0 {
			return float64(stacks)
		}
	}
	return 1
}

func (p *Parsed) set(sim *core.Simulation, level float64) {
	for _, a := range p.attachments {
		if a != nil {
			a.set(sim, level)
		}
	}
}

// What core's Attach helpers do for an aura that is already up when they are attached: apply now,
// with no Simulation in hand. The rows that read one are left for the aura's next application.
func (p *Parsed) catchUp(level float64) {
	for _, a := range p.attachments {
		if a != nil && !a.needsSim {
			a.set(nil, level)
		}
	}
}

func (p *Parsed) skip(s *Spell, i int, e *Effect) {
	p.Skipped = append(p.Skipped, e)
	report(s, i+1, e)
}

// Whether the parse reads the effect at this position.
func (o *parseOptions) reads(pos int32) bool {
	if len(o.only) > 0 && !slices.Contains(o.only, pos) {
		return false
	}
	return !slices.Contains(o.skip, pos)
}

// The effect types that put an aura on someone: the plain application and the area auras, which
// carry the same aura and misc values.
func appliesAura(t dbcenums.SpellEffectType) bool {
	return t == dbcenums.E_APPLY_AURA || t == dbcenums.E_APPLY_AREA_AURA_PARTY ||
		t == dbcenums.E_APPLY_AREA_AURA_RAID
}

// The dot modifiers a spell states twice. A talent that raises a spell's damage states the same
// number once for the hit and once for the dot, and one SpellMod_DamageDone_Flat already reaches the
// ticks as well as the hit, so the second effect would double the bonus on the dot. Only a hit
// modifier this parse reads folds the dot one into itself; a parse narrowed to the dot effect alone
// attaches it.
func foldedDotEffects(s *Spell, o *parseOptions) []*Effect {
	var folded []*Effect
	for i := range s.Effects {
		dot := &s.Effects[i]
		if dot.Aura != dbcenums.A_ADD_PCT_MODIFIER || dot.Misc != SPELLMOD_DOT {
			continue
		}

		for j := range s.Effects {
			hit := &s.Effects[j]
			if hit.Aura != dbcenums.A_ADD_PCT_MODIFIER || !appliesAura(hit.Type) ||
				!o.reads(int32(j+1)) ||
				(hit.Misc != SPELLMOD_DAMAGE && hit.Misc != SPELLMOD_ALL_EFFECTS) {
				continue
			}
			if hit.ClassFlags == dot.ClassFlags && hit.BasePoints == dot.BasePoints {
				folded = append(folded, dot)
				break
			}
		}
	}
	return folded
}

// What a port has to wire by hand, on the console when SPELLDATA_REPORT is set.
func report(s *Spell, pos int, e *Effect) {
	if os.Getenv("SPELLDATA_REPORT") == "" {
		return
	}
	fmt.Printf("spelldata: unparsed %d %s effect %d %s(%d) misc %d value %v\n",
		s.ID, s.Name, pos, auraName(e.Aura), e.Aura, e.Misc, e.BasePoints)
}

// The client's name for an aura the parser skipped. Only the auras a modifier or a buff row is
// expected to carry are named; anything else reads as its number.
func auraName(a dbcenums.EffectAuraType) string {
	if name, ok := auraNames[a]; ok {
		return name
	}
	return fmt.Sprintf("A_%d", a)
}

var auraNames = map[dbcenums.EffectAuraType]string{
	dbcenums.A_PERIODIC_DAMAGE:                  "A_PERIODIC_DAMAGE",
	dbcenums.A_DUMMY:                            "A_DUMMY",
	dbcenums.A_MOD_FEAR:                         "A_MOD_FEAR",
	dbcenums.A_PERIODIC_HEAL:                    "A_PERIODIC_HEAL",
	dbcenums.A_MOD_ATTACKSPEED:                  "A_MOD_ATTACKSPEED",
	dbcenums.A_MOD_THREAT:                       "A_MOD_THREAT",
	dbcenums.A_MOD_TAUNT:                        "A_MOD_TAUNT",
	dbcenums.A_MOD_STUN:                         "A_MOD_STUN",
	dbcenums.A_MOD_DAMAGE_DONE:                  "A_MOD_DAMAGE_DONE",
	dbcenums.A_MOD_DAMAGE_TAKEN:                 "A_MOD_DAMAGE_TAKEN",
	dbcenums.A_DAMAGE_SHIELD:                    "A_DAMAGE_SHIELD",
	dbcenums.A_MOD_RESISTANCE:                   "A_MOD_RESISTANCE",
	dbcenums.A_PERIODIC_TRIGGER_SPELL:           "A_PERIODIC_TRIGGER_SPELL",
	dbcenums.A_PERIODIC_ENERGIZE:                "A_PERIODIC_ENERGIZE",
	dbcenums.A_MOD_ROOT:                         "A_MOD_ROOT",
	dbcenums.A_MOD_SILENCE:                      "A_MOD_SILENCE",
	dbcenums.A_MOD_STAT:                         "A_MOD_STAT",
	dbcenums.A_MOD_INCREASE_SPEED:               "A_MOD_INCREASE_SPEED",
	dbcenums.A_MOD_DECREASE_SPEED:               "A_MOD_DECREASE_SPEED",
	dbcenums.A_MOD_INCREASE_HEALTH:              "A_MOD_INCREASE_HEALTH",
	dbcenums.A_MOD_SHAPESHIFT:                   "A_MOD_SHAPESHIFT",
	dbcenums.A_SCHOOL_IMMUNITY:                  "A_SCHOOL_IMMUNITY",
	dbcenums.A_PROC_TRIGGER_SPELL:               "A_PROC_TRIGGER_SPELL",
	dbcenums.A_MOD_PARRY_PERCENT:                "A_MOD_PARRY_PERCENT",
	dbcenums.A_MOD_DODGE_PERCENT:                "A_MOD_DODGE_PERCENT",
	dbcenums.A_MOD_BLOCK_PERCENT:                "A_MOD_BLOCK_PERCENT",
	dbcenums.A_MOD_WEAPON_CRIT_PERCENT:          "A_MOD_WEAPON_CRIT_PERCENT",
	dbcenums.A_MOD_HIT_CHANCE:                   "A_MOD_HIT_CHANCE",
	dbcenums.A_MOD_SPELL_HIT_CHANCE:             "A_MOD_SPELL_HIT_CHANCE",
	dbcenums.A_TRANSFORM:                        "A_TRANSFORM",
	dbcenums.A_MOD_SPELL_CRIT_CHANCE:            "A_MOD_SPELL_CRIT_CHANCE",
	dbcenums.A_MOD_CASTING_SPEED_NOT_STACK:      "A_MOD_CASTING_SPEED_NOT_STACK",
	dbcenums.A_MECHANIC_IMMUNITY:                "A_MECHANIC_IMMUNITY",
	dbcenums.A_MOD_DAMAGE_PERCENT_DONE:          "A_MOD_DAMAGE_PERCENT_DONE",
	dbcenums.A_MOD_POWER_REGEN:                  "A_MOD_POWER_REGEN",
	dbcenums.A_MOD_DAMAGE_PERCENT_TAKEN:         "A_MOD_DAMAGE_PERCENT_TAKEN",
	dbcenums.A_MOD_ATTACK_POWER:                 "A_MOD_ATTACK_POWER",
	dbcenums.A_ADD_FLAT_MODIFIER:                "A_ADD_FLAT_MODIFIER",
	dbcenums.A_ADD_PCT_MODIFIER:                 "A_ADD_PCT_MODIFIER",
	dbcenums.A_ADD_TARGET_TRIGGER:               "A_ADD_TARGET_TRIGGER",
	dbcenums.A_MOD_HEALING:                      "A_MOD_HEALING",
	dbcenums.A_MOD_HEALING_PCT:                  "A_MOD_HEALING_PCT",
	dbcenums.A_MOD_OFFHAND_DAMAGE_PCT:           "A_MOD_OFFHAND_DAMAGE_PCT",
	dbcenums.A_MOD_RANGED_ATTACK_POWER:          "A_MOD_RANGED_ATTACK_POWER",
	dbcenums.A_MOD_INCREASE_HEALTH_PERCENT:      "A_MOD_INCREASE_HEALTH_PERCENT",
	dbcenums.A_MOD_HEALING_DONE:                 "A_MOD_HEALING_DONE",
	dbcenums.A_MOD_HEALING_DONE_PERCENT:         "A_MOD_HEALING_DONE_PERCENT",
	dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE:        "A_MOD_TOTAL_STAT_PERCENTAGE",
	dbcenums.A_MOD_BASE_RESISTANCE_PCT:          "A_MOD_BASE_RESISTANCE_PCT",
	dbcenums.A_MOD_CRIT_DAMAGE_BONUS:            "A_MOD_CRIT_DAMAGE_BONUS",
	dbcenums.A_OVERRIDE_CLASS_SCRIPTS:           "A_OVERRIDE_CLASS_SCRIPTS",
	dbcenums.A_MOD_IGNORE_SHAPESHIFT:            "A_MOD_IGNORE_SHAPESHIFT",
	dbcenums.A_MECHANIC_DURATION_MOD:            "A_MECHANIC_DURATION_MOD",
	dbcenums.A_MOD_EXPERTISE:                    "A_MOD_EXPERTISE",
	dbcenums.A_MOD_CRIT_PCT:                     "A_MOD_CRIT_PCT",
	dbcenums.A_MOD_MELEE_HASTE_3:                "A_MOD_MELEE_HASTE_3",
	dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS:        "A_OVERRIDE_ACTIONBAR_SPELLS",
	dbcenums.A_MOD_ADDITIONAL_POWER_COST:        "A_MOD_ADDITIONAL_POWER_COST",
	dbcenums.A_MOD_POWER_COST_SCHOOL_PCT:        "A_MOD_POWER_COST_SCHOOL_PCT",
	dbcenums.A_MOD_INCREASE_ENERGY:              "A_MOD_INCREASE_ENERGY",
	dbcenums.A_MOD_SKILL:                        "A_MOD_SKILL",
	dbcenums.A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT: "A_MOD_SPELL_DAMAGE_OF_STAT_PERCENT",
}
