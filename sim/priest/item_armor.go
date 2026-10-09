package priest

func init() {
	// The All-Seeing Eye of Zuldazar
	// https://www.wowhead.com/forever/item=19594/the-all-seeing-eye-of-zuldazar
	//
	// Increases the amount of damage absorbed by Power Word: Shield by 35.
	//
	// TODO: 24191 states the 35 as a flat SPELLMOD_ALL_EFFECTS, which the parse has no row for, so
	// the aura would carry nothing even once Power Word: Shield is ported.
	// shared.NewSpellDataEquipAura(shared.SpellDataProc{TriggerSpellID: 24191},
	// 	[]shared.ItemVariant{
	// 		{ItemID: 19594, ItemName: "The All-Seeing Eye of Zuldazar"},
	// 	})
}
