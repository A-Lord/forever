package overrides

// Which roll an EnchantProc's Value is.
type RateKind uint8

const (
	RatePPM    RateKind = iota + 1 // Procs per minute, measured against the enchanted weapon's speed.
	RateChance                     // A chance per hit, 0 to 1.
)

func (k RateKind) String() string {
	switch k {
	case RatePPM:
		return "ppm"
	case RateChance:
		return "chance"
	}
	return "unknown rate kind"
}

// A permanent enchant's proc rate, which the game files keep outside the spell data: the slot's spell
// states no chance, or the 100 or 101 sentinel. The generator puts it onto the enchant's effect for
// the spell, where the sim reads it. Reason and Source weigh the same as an Override's.
type EnchantProc struct {
	EnchantID int32
	SpellID   int32
	Kind      RateKind
	Value     float64
	Reason    string
	Source    string // "tbc-carryover", "vanilla-carryover", "wcl:<report>/<fight>", "issue #N"
}

var EnchantProcs = []EnchantProc{
	{
		EnchantID: 1900,
		SpellID:   20007,
		Kind:      RatePPM,
		Value:     1, // TODO: TBC value - verify in-game
		Reason:    "Crusader: 20007 states no proc chance, and 458112's 100 is the sentinel beside 'often'",
		Source:    "tbc-carryover",
	},
}

// The rate stated for an enchant's proc spell, or nil where none is.
func EnchantProcRate(enchantID int32, spellID int32) *EnchantProc {
	for i := range EnchantProcs {
		if EnchantProcs[i].EnchantID == enchantID && EnchantProcs[i].SpellID == spellID {
			return &EnchantProcs[i]
		}
	}
	return nil
}
