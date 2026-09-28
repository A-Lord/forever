package overrides

// A spell the store carries although no class, item, enchant or set bonus reaches it. Reason and
// Source weigh the same as an Override's, and the generator refuses an entry without a reason.
type Extra struct {
	SpellID int32
	Reason  string
	Source  string
}

// The spells the generator adds to the store's roots by hand.
var ExtraSpells = []Extra{
	{SpellID: 1293697, Reason: "Demonic Brand's Shadow damage when a demon consumes the brand; nothing links it to the talent", Source: "Spell 1293697 description"},
	{SpellID: 1293698, Reason: "Demonic Brand's Fire damage when a demon consumes the brand; nothing links it to the talent", Source: "Spell 1293698 description"},
}
