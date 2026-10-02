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
	// Enchant Weapon - Revelation 8217: its proc spell 1248806 names no trigger, and each class with
	// spells to crit has a buff of its own, which the sim picks by the wearer's class.
	{SpellID: 1248808, Reason: "Revelation, the mage buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323377, Reason: "Revelation, the priest buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323392, Reason: "Revelation, the warlock buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323400, Reason: "Revelation, the druid buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323410, Reason: "Revelation, the hunter buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323418, Reason: "Revelation, the shaman buff of enchant 8217", Source: "owner decision"},
	{SpellID: 1323419, Reason: "Revelation, the paladin buff of enchant 8217", Source: "owner decision"},
}
