package druid

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var druidSetSlots = []proto.ItemSlot{
	proto.ItemSlot_ItemSlotHead, proto.ItemSlot_ItemSlotShoulder,
	proto.ItemSlot_ItemSlotChest, proto.ItemSlot_ItemSlotWrist,
	proto.ItemSlot_ItemSlotHands, proto.ItemSlot_ItemSlotWaist,
	proto.ItemSlot_ItemSlotLegs, proto.ItemSlot_ItemSlotFeet,
}

func druidSetBonus(spellIDs ...int32) core.ApplySetBonus {
	return func(agent core.Agent, aura *core.Aura) {
		for _, id := range spellIDs {
			spelldata.ParseEffects(agent.GetCharacter(), aura, spelldata.MustFind(id))
		}
	}
}

var ItemSetWildheartRaiment = core.NewItemSet(core.ItemSet{
	Name:  "Wildheart Raiment",
	ID:    185,
	Slots: druidSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: druidSetBonus(18679),
		3: druidSetBonus(9346, 9336),
		// TODO: The 4-piece movement proc and 5-piece resource proc are not modelled.
		6: druidSetBonus(18378),
	},
})

var ItemSetFeralheartRaiment = core.NewItemSet(core.ItemSet{
	Name:  "Feralheart Raiment",
	ID:    513,
	Slots: druidSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: druidSetBonus(18679),
		// TODO: The 4-piece resource proc's tooltip and referenced energy and mana rows disagree.
		6: druidSetBonus(9344, 9334),
		8: druidSetBonus(14803),
	},
})

var ItemSetGenesisRaiment = core.NewItemSet(core.ItemSet{
	Name:  "Genesis Raiment",
	ID:    493,
	Slots: druidSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		3: func(agent core.Agent, aura *core.Aura) {
			druidSetBonus(15693)(agent, aura)
			aura.AttachStatBuff(stats.DefenseRating, spelldata.MustFind(18196).EffectN(1).BaseValue()*core.DefenseRatingPerDefenseLevel)
		},
		5: druidSetBonus(26106),
	},
})

var ItemSetGrovekeeperFerocity = core.NewItemSet(core.ItemSet{
	Name:  "Grovekeeper Ferocity",
	ID:    2115,
	Slots: druidSetSlots,
	Bonuses: map[int32]core.ApplySetBonus{
		2: druidSetBonus(1300934),
		4: druidSetBonus(1301075),
		5: druidSetBonus(1301247),
	},
})
