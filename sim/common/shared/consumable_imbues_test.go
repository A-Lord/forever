package shared

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Brilliant Wizard Oil 20749, enchant 2628: equip spell 25113's 36 healing, 36 spell damage and 1%
// spell crit, once for each hand carrying it.
func TestTemporaryEnchantStatsApplyPerEnchantedHand(t *testing.T) {
	oil := core.Consumable{Id: 20749, Name: "Brilliant Wizard Oil", TemporaryEnchant: &core.Enchant{
		EffectID:    2628,
		Stats:       stats.Stats{stats.HealingPower: 36, stats.SpellDamage: 36},
		PseudoStats: pseudoStats(proto.PseudoStat_PseudoStatSpellCritPercent, 1),
	}}

	character := newTestAgent().character
	character.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.Item{ID: 1, TempEnchant: 2628}
	before := character.GetStats()
	applyTemporaryEnchant(character, oil)
	gained := character.GetStats().Subtract(before)

	if gained[stats.SpellDamage] != 36 || gained[stats.HealingPower] != 36 || gained[stats.SpellCritPercent] != 1 {
		t.Errorf("one oiled hand gained %v spell damage, %v healing, %v%% spell crit, want 36, 36 and 1",
			gained[stats.SpellDamage], gained[stats.HealingPower], gained[stats.SpellCritPercent])
	}

	bare := newTestAgent().character
	before = bare.GetStats()
	applyTemporaryEnchant(bare, oil)
	if gained := bare.GetStats().Subtract(before); gained != (stats.Stats{}) {
		t.Errorf("a character with no hand carrying the enchant gained %v", gained)
	}
}

// Dense Sharpening Stone 12404, enchant 1643: 8 weapon damage on the hand carrying it only.
func TestTemporaryEnchantWeaponDamageGoesOnItsHand(t *testing.T) {
	stone := core.Consumable{Id: 12404, Name: "Dense Sharpening Stone",
		TemporaryEnchant: &core.Enchant{EffectID: 1643, WeaponDamage: 8}}

	character := newTestAgent().character
	character.Equipment[proto.ItemSlot_ItemSlotOffHand] = core.Item{ID: 1, TempEnchant: 1643, SwingSpeed: 2.6}
	*character.AutoAttacks.MH() = core.Weapon{BaseDamageMin: 100, BaseDamageMax: 150, SwingSpeed: 2.6}
	*character.AutoAttacks.OH() = core.Weapon{BaseDamageMin: 100, BaseDamageMax: 150, SwingSpeed: 2.6}
	applyTemporaryEnchant(character, stone)

	if oh := character.AutoAttacks.OH(); oh.BaseDamageMin != 108 || oh.BaseDamageMax != 158 {
		t.Errorf("off hand deals %v-%v, want 108-158", oh.BaseDamageMin, oh.BaseDamageMax)
	}
	if mh := character.AutoAttacks.MH(); mh.BaseDamageMin != 100 || mh.BaseDamageMax != 150 {
		t.Errorf("main hand deals %v-%v, want its own 100-150", mh.BaseDamageMin, mh.BaseDamageMax)
	}
}

// Shadow Oil 3824, enchant 25: its combat spell 350026 is cast off the enchanted weapon's hits.
func TestTemporaryEnchantCombatSpellRegistersAWeaponProc(t *testing.T) {
	shadowOil := core.Consumable{Id: 3824, Name: "Shadow Oil", TemporaryEnchant: &core.Enchant{
		EffectID:       25,
		EnchantEffects: []*proto.ItemEffect{{BuffId: 350026, Effect: &proto.ItemEffect_Proc{Proc: &proto.ProcEffect{}}}},
	}}

	character := newTestAgent().character
	character.Equipment[proto.ItemSlot_ItemSlotMainHand] = core.Item{ID: 1, TempEnchant: 25}
	applyTemporaryEnchant(character, shadowOil)

	if auraByLabel(character, "Shadow Oil") == nil {
		t.Fatal("no listener casts the oil's combat spell")
	}
	if character.GetSpell(core.ActionID{SpellID: 350026}) == nil {
		t.Error("the combat spell 350026 was not registered")
	}
}

func pseudoStats(stat proto.PseudoStat, value float64) []float64 {
	pseudo := make([]float64, stats.PseudoStatsLen)
	pseudo[stat] = value
	return pseudo
}
