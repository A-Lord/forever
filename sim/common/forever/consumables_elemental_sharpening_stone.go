package forever

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Elemental Sharpening Stone 18262 puts enchant 2506 on a weapon, whose equip spell 22755 states
// A_MOD_WEAPON_CRIT_PERCENT 2 on the melee weapons its SpellEquippedItems mask 42483 names. Each
// weapon carries its own copy of the aura, so the crit reaches the attacks made with the hand the
// stone is on, while that hand holds a weapon the mask names, and a magic-defense spell off that
// hand gains none.
const elementalSharpeningStone = 18262

func init() {
	core.NewConsumableEffect(elementalSharpeningStone, func(agent core.Agent) {
		character := agent.GetCharacter()
		enchant := core.GetConsumableByID(elementalSharpeningStone).TemporaryEnchant
		if enchant == nil {
			return
		}
		row := spelldata.MustFind(22755)
		crit := row.EffectN(1).Average(character.Level)

		for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
			if character.Equipment[slot].TempEnchant != enchant.EffectID {
				continue
			}
			hand := character.GetProcMaskForWeaponSlot(slot)
			character.AddStaticMod(core.SpellModConfig{
				Kind:          core.SpellMod_BonusCrit_Percent,
				FloatValue:    crit,
				ProcMask:      hand,
				EquipHand:     hand,
				EquipSubclass: row.EquipSubclass,
			})
		}
	})
}
