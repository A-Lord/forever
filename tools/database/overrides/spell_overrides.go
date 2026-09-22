// Package overrides holds the numbers the spelldata store carries that the client database does
// not state. Each row names the spell, the field, the value and where the value comes from, and the
// generator refuses to run when the client starts stating one itself - a value kept by hand after
// the data grew a column for it is a wrong number nobody would notice.
//
// A leaf package on purpose: tools/database imports it, it imports nothing back, and the folder it
// shares with the SQL overrides is read for *.sql alone.
package overrides

// Which field of the row the value belongs on. Every one of them names a rule the generator checks
// before it writes the value, so an override outlives its reason no longer than the next
// regeneration.
type Field uint8

const (
	// Procs per minute, onto Spell.RPPM. Fails when SpellAuraOptions.SpellProcsPerMinuteID stops
	// being zero on the row.
	PPM Field = iota + 1

	// A flat threat bonus, onto Spell.FlatThreat. Fails when the spell gains an E_THREAT or
	// E_THREAT_ALL effect.
	FlatThreat

	// The attack-power coefficient of the direct effect, onto Effect.APCoef. Fails when the client
	// states a BonusCoefficientFromAP there.
	APCoefDirect

	// The same for the ticking effect.
	APCoefPeriodic

	// A proc chance in percent, onto Spell.ProcChance. Fails when the tooltip states a chance of
	// its own through $h or $mN, because then the client already says where the roll is.
	ProcChancePct

	// The aura's duration in milliseconds. Fails when the client states a SpellDuration.
	DurationMs

	// The ProcHint bits the tooltip's wording did not yield, which Value carries as a bit set. It
	// replaces what the wording produced rather than adding to it.
	Hint
)

func (f Field) String() string {
	switch f {
	case PPM:
		return "PPM"
	case FlatThreat:
		return "FlatThreat"
	case APCoefDirect:
		return "APCoefDirect"
	case APCoefPeriodic:
		return "APCoefPeriodic"
	case ProcChancePct:
		return "ProcChancePct"
	case DurationMs:
		return "DurationMs"
	case Hint:
		return "Hint"
	}
	return "unknown field"
}

// One hand-supplied value. Reason and Source are not documentation: they are what the next reader
// has to weigh the number against, so the generator refuses a row without a reason.
type Override struct {
	SpellID int32
	Field   Field
	Value   float64
	Reason  string
	Source  string // "tbc-carryover", "tooltip", "wcl:<report>/<fight>", "issue #N"
}

// The item procs whose rate the client keeps outside the spell data. Their ProcChance column reads
// the 100 or 101 sentinel, which the store would otherwise take for "fires on every hit", and the
// rates are the ones tools/database/dbc/maps.go MapItemIdToPPM carries for the same items - the
// three of that list Forever ships. None has been measured on this server.
var Spells = []Override{
	{16928, PPM, 1, "Annihilator: ProcChance 101 sentinel; TBC 1 PPM, unverified on Forever", "tbc-carryover"},
	{23686, PPM, 1, "Darkmoon Card: Maelstrom: column reads 100, tooltip says 'Chance to strike'", "tbc-carryover"},
	{26480, PPM, 10, "Badge of the Swarmguard: stack accumulator inside the on-use window", "tbc-carryover"},
}

// The procs-per-minute rate stated for a spell, or zero where none is. The item database reads it
// for the same spells the store does: a container whose ProcChance column is the "rate lives
// elsewhere" sentinel has no rate anywhere else either.
func PPMFor(spellID int32) float64 {
	for _, override := range Spells {
		if override.SpellID == spellID && override.Field == PPM {
			return override.Value
		}
	}
	return 0
}
