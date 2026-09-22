package forever

func RegisterAllEnchants() {

	// Enchants

	// TODO: Manual implementation required
	//       This can be ignored if the effect has already been implemented.
	//       With next db run the item will be removed if implemented.
	//
	// Permanently enchant a melee weapon to often inflict a curse on the target reducing their melee damage.
	// https://www.wowhead.com/forever/spell=20033
	// unsupported: states no rate
	// trigger 20006 (0%, core.CallbackEmpty, core.ProcMaskUnknown)
	// shared.NewSpellDataProc(shared.SpellDataProc{
	//	Name:           "Enchant Weapon - Unholy Weapon",
	//	EnchantID:      1899,
	//	TriggerSpellID: 20006,
	//	IsWeaponProc:   true,
	// }, nil)
}
