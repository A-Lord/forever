package spelldata

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
)

// An addition to the resolved trigger for what the client does not state, which on a proc is the
// rate: a row that states none needs one of PPM, ChanceFrom or Chance.
type ProcOpt func(*core.Character, *core.ProcTrigger)

// What the client states about a proc, as the fields core registers a listener through: what it
// hears (ProcTypeMask through the decoder, with the row's tooltip hint), how often it fires, its
// internal cooldown and the attribute-driven gates. The caller adds what the row does not carry -
// the name and action the sim keys the effect by, the buff the handler applies - to the returned
// value before handing it to MakeProcTriggerAura.
//
// ProcCharges is not a property of the trigger: it is how many times the buff acts before it drops,
// which AuraConfig reads as the aura's MaxStacks.
//
// The character is needed for the procs-per-minute managers, which are bound to the weapon carrying
// them, so a trigger is built where the sim has a character rather than at package init.
func ProcTrigger(character *core.Character, s *Spell, handler core.ProcHandler, opts ...ProcOpt) core.ProcTrigger {
	decoded := core.DecodeProcTypeMask(s.ProcFlags, s.ProcHint)

	trigger := core.ProcTrigger{
		Name:               s.Name,
		ActionID:           core.ActionID{SpellID: s.ID},
		Callback:           decoded.Callback,
		ProcMask:           decoded.ProcMask,
		Outcome:            decoded.Outcome,
		RequireDamageDealt: decoded.RequireDamageDealt,
		ICD:                s.ICD(),
		CanProcFromProcs:   s.CanProcFromProcs(),
		ClassSpellsOnly:    s.ClassSpellsOnly(),
		ClassFlags:         procClassFlags(character, s),
		Handler:            handler,
	}

	if s.IsWeaponProcAura() {
		trigger.SpellFlagsExclude |= core.SpellFlagSuppressWeaponProcs
	}

	fillProcChance(s, &trigger)

	// The row first and the caller's options on top, so an option sees the mask the row decoded to
	// and can override a rate the row states wrongly or not at all.
	for _, opt := range opts {
		opt(character, &trigger)
	}

	settleProcRate(character, s, &trigger)

	return trigger
}

// A hand-supplied rate, for a proc whose real rate lives outside the spell data. It reads the
// trigger's own proc mask, so an option that narrows the mask has to be given before this one.
func PPM(ppm float64) ProcOpt {
	return func(character *core.Character, trigger *core.ProcTrigger) {
		// The callback rolls the chance first and the manager only where the chance let it
		// through, so a column chance left next to a manager would gate the rate twice.
		trigger.ProcChance = 0
		trigger.DPM = character.NewLegacyPPMManager(ppm, trigger.ProcMask)
	}
}

// The chance an effect states, for a $mN the generator could not resolve into ProcChanceEffect.
func ChanceFrom(e *Effect) ProcOpt {
	return Chance(e.Percent())
}

// A rate the caller states outright, which replaces whatever the row said, manager included.
func Chance(chance float64) ProcOpt {
	return func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = chance
		trigger.DPM = nil
	}
}

// What a trigger built from this row does not model: the proc flags the decoder names as
// unsupported, the two tooltip shapes it reads past, and a class mask naming a family the wearer's
// class does not use. A trigger is still built for all of them - a listener that hears fewer hits
// than the client's is deliberately the narrower one - so this is the audit's way of seeing which
// rows that applies to.
func ProcTriggerUnsupported(character *core.Character, s *Spell) []string {
	unsupported := core.DecodeProcTypeMask(s.ProcFlags, s.ProcHint).Unsupported

	// No ProcTypeMask can state either: a trigger restricted to one named ability, and one
	// restricted to an outcome the mask has no bit for, both listen to more than the client does.
	if s.ProcHint.Matches(core.ProcHintNamedAbility) {
		unsupported = append(unsupported, "NAMED_ABILITY")
	}
	if s.ProcHint.Matches(core.ProcHintOutcomeTaken) {
		unsupported = append(unsupported, "OUTCOME_TAKEN")
	}
	if othersFamily(character, rowClassFlags(s)) {
		unsupported = append(unsupported, "CLASS_MASK_OTHER_FAMILY")
	}

	return unsupported
}

// The spells the listener fires on where the client names them, which is the EffectSpellClassMask of
// the effect that carries the proc, read against the class wearing it.
func procClassFlags(character *core.Character, s *Spell) core.ClassFlags {
	flags := rowClassFlags(s)
	if othersFamily(character, flags) {
		return core.ClassFlags{}
	}

	return flags
}

// Only the proc effect's flags: another effect of the same spell states the spells it modifies,
// not the ones that feed the proc.
func rowClassFlags(s *Spell) core.ClassFlags {
	for i := range s.Effects {
		e := &s.Effects[i]

		switch e.Aura {
		case dbcenums.A_PROC_TRIGGER_SPELL, dbcenums.A_PROC_TRIGGER_SPELL_WITH_VALUE,
			dbcenums.A_PROC_TRIGGER_SPELL_COPY, dbcenums.A_PROC_TRIGGER_DAMAGE, dbcenums.A_DUMMY:
			if !e.ClassFlags.IsZero() {
				return e.ClassFlags
			}
		}
	}

	return core.ClassFlags{}
}

// Whether the mask names spells of a family this character's class never casts, which is how an
// item shared by every class states the filter of the one class it was written for: Dreadnaught's
// 8pc 28845 names family 5, the warlock's, and on a warrior that mask matches nothing and silences
// the listener rather than narrowing it. A class the client states no family for - and the empty
// mask every non-class spell carries - is no evidence of anything, so both are left alone.
//
// Only another class's family counts. The client files the potions under family 13 and a good deal
// of generic content under 0 or 1, and a mask in one of those names spells every class can use, so
// dropping it would silence a listener the wearer really does hear. With no character there is no
// class to read the mask against, which ParseEffects answers the same way: the rows that need one
// are left out rather than guessed at.
func othersFamily(character *core.Character, flags core.ClassFlags) bool {
	if character == nil {
		return false
	}

	family, stated := classSpellFamilies[character.Class]
	return stated && !flags.IsZero() && flags.Family != family && isClassFamily(flags.Family)
}

// Whether the family is the one some class files its own spells under.
func isClassFamily(family int32) bool {
	for _, classFamily := range classSpellFamilies {
		if classFamily == family {
			return true
		}
	}
	return false
}

// The client's SpellClassSet per class: the family every one of that class's spells files its class
// mask under. Verified row by row against the store in TestClassSpellFamilies.
var classSpellFamilies = map[proto.Class]int32{
	proto.Class_ClassMage:    3,
	proto.Class_ClassWarrior: 4,
	proto.Class_ClassWarlock: 5,
	proto.Class_ClassPriest:  6,
	proto.Class_ClassDruid:   7,
	proto.Class_ClassRogue:   8,
	proto.Class_ClassHunter:  9,
	proto.Class_ClassPaladin: 10,
	proto.Class_ClassShaman:  11,
}

// The roll the row states, by the source that says where it is stated. The ProcChance column is the
// roll only under ProcChanceColumn; under the others it means nothing, which is why reading it
// directly is the bug ProcChanceSource exists to prevent.
//
// A row whose rate is procs per minute states no roll: that rate is a manager, and settleProcRate
// builds it once the options have had their say about the mask it measures hits on.
func fillProcChance(s *Spell, trigger *core.ProcTrigger) {
	if s.RPPM > 0 {
		return
	}

	switch s.ProcChanceSource {
	case ProcChanceColumn:
		// 101 is the client's "fires on its own condition" sentinel next to a real 100.
		trigger.ProcChance = min(float64(s.ProcChance)/100, 1)
	case ProcChanceEffectN:
		trigger.ProcChance = s.EffectN(int(s.ProcChanceEffect)).Percent()
	case ProcChanceAlways:
		trigger.ProcChance = 1
	case ProcChancePPM:
		// The client states nothing, so the rate is the caller's to supply through PPM().
	}
}

// The rate the trigger ends up with. An override-supplied procs-per-minute rate wins over every
// column - the rows that carry one are the rows whose stated chance the tooltip contradicts - but it
// is measured against the mask the trigger *ends* on rather than the one the row decoded to: an
// option that narrows or blanks the mask, as a weapon proc's shape does, would otherwise be handed a
// manager counting hits the listener no longer hears.
//
// A trigger left with neither a chance nor a manager fires on every qualifying hit, which is never
// what a row with no stated rate means.
func settleProcRate(character *core.Character, s *Spell, trigger *core.ProcTrigger) {
	if trigger.ProcChance != 0 || trigger.DPM != nil {
		return
	}

	if s.RPPM > 0 {
		// A mask of nothing is the weapon-proc shape, where which hits count is the weapon's to say
		// and the manager has to be bound to it by whoever knows which weapon that is.
		if trigger.ProcMask == core.ProcMaskUnknown {
			panic(fmt.Sprintf("spelldata: spell %d (%s) states %g procs per minute but no proc mask to measure them on; pass a manager bound to what carries it",
				s.ID, s.Name, s.RPPM))
		}

		trigger.DPM = character.NewLegacyPPMManager(float64(s.RPPM), trigger.ProcMask)
		return
	}

	panic(fmt.Sprintf("spelldata: spell %d (%s) states no proc chance; pass PPM()", s.ID, s.Name))
}
