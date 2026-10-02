package overrides

// A spell a server-side handler casts off another, which no effect edge, $<id> token or skill-line
// row names. Reason and Source weigh the same as an Override's, and the generator refuses an entry
// without a reason.
type HandTrigger struct {
	Spell    int32
	Triggers int32
	Reason   string
	Source   string
}

// The trigger edges the generator adds by hand to the store.
var HandTriggers = []HandTrigger{
	{
		Spell:    20230,
		Triggers: 20240,
		Reason: "Retaliation's dummy aura (aura 4) casts the counterattack 20240: same name, class set and " +
			"icon, weapon damage with no base, a cost of 1 in SpellPower that is a tenth of a rage",
		Source: "game-files-rows",
	},
	{
		Spell:    16344,
		Triggers: 10444,
		Reason: "Flametongue Weapon Proc rank 6 states no spell power share; Flametongue Attack 10444 carries it: " +
			"same class set and mask 0x200000, fire school and icon, a school damage effect of base 1 with a 0.1 coefficient",
		Source: "game-files-rows",
	},
}
