package dbcenums

// Spell attribute flags, named after the Attributes column they live in: ATTR_EX_3 is a flag
// in Attributes[3]. Read them through the Spell helpers rather than indexing Attributes.
const (
	// The spell is never cast: a stance's passive, a talent that only modifies other spells.
	ATTR_PASSIVE uint32 = 0x40

	// The two bits the client marks a channel with. Arcane Missiles and Blizzard carry the first,
	// Evocation and Tranquility only the second, so a channel check has to read both.
	ATTR_EX_1_IS_CHANNELLED      uint32 = 0x4
	ATTR_EX_1_IS_SELF_CHANNELLED uint32 = 0x40

	// The server refunds 80% of the power cost when the spell misses: every rage special that
	// costs rage up front, Heroic Strike and Rend among them. Cleave and Whirlwind lack it.
	ATTR_EX_1_DISCOUNT_POWER_ON_MISS uint32 = 0x8000000

	ATTR_EX_2_CANT_CRIT uint32 = 0x20000000

	// On a triggered spell: aura listeners treat its hits like a normal ability hit. Seal of
	// Command damage, every Judgement, Stormstrike's bonus hits and Sweeping Strikes carry it.
	ATTR_EX_3_NOT_A_PROC          uint32 = 0x200
	ATTR_EX_3_CAN_PROC_FROM_PROCS uint32 = 0x4000000

	// Weapon procs (Player::CastItemCombatSpell) ignore hits of this spell, as do auras marked
	// ATTR_EX_6_AURA_IS_WEAPON_PROC. In TBC that is the Seal of Blood, Righteousness and Martyr
	// damage spells plus Gouge, Sap, Scatter Shot and Maim.
	ATTR_EX_4_SUPPRESS_WEAPON_PROCS uint32 = 0x800000

	// An aura proc that honours ATTR_EX_4_SUPPRESS_WEAPON_PROCS anyway: Black Bow of the Betrayer
	// and the Sunwell melee neck.
	ATTR_EX_6_AURA_IS_WEAPON_PROC uint32 = 0x80

	// A periodic effect whose ticks roll a critical strike: Rend, Corruption, Rupture and the
	// other bleeds and DoTs the client marks, 225 spells in this build.
	ATTR_EX_8_PERIODIC_CAN_CRIT uint32 = 0x200

	ATTR_EX_11_SCALES_WITH_ITEM_LEVEL uint32 = 0x4

	ATTR_EX_12_ONLY_PROC_FROM_CLASS_ABILITIES uint32 = 0x80000000
)

// Attributes index each ATTR_EX_ flag above belongs to.
const (
	ATTR_INDEX_BASE  int = 0
	ATTR_INDEX_EX_1  int = 1
	ATTR_INDEX_EX_2  int = 2
	ATTR_INDEX_EX_3  int = 3
	ATTR_INDEX_EX_4  int = 4
	ATTR_INDEX_EX_6  int = 6
	ATTR_INDEX_EX_8  int = 8
	ATTR_INDEX_EX_11 int = 11
	ATTR_INDEX_EX_12 int = 12
)
