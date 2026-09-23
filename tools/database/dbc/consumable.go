package dbc

import (
	"slices"

	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

type Consumable struct {
	Id                       int             // Item ID
	Name                     string          // Item name
	ItemLevel                int             // Item level
	RequiredLevel            int             // Required level to use
	ClassId                  int             // Item class ID (should be 0 for consumables)
	SubClassId               ConsumableClass // Item subclass ID
	IconFileDataID           int             // Icon file data ID
	SpellCategoryID          int             // Spell category ID
	SpellCategoryFlags       int             // Spell category flags
	ItemEffects              []int           // Item effect IDs
	ElixirType               int
	Duration                 int // In milliseconds
	CooldownDuration         int // In milliseconds
	CategoryCooldownDuration int // In milliseconds
	ClassMask                int
	RaceMask                 int64
	RequiredSkill            int
}

func (c *Consumable) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"Id":                 c.Id,
		"Name":               c.Name,
		"ItemLevel":          c.ItemLevel,
		"RequiredLevel":      c.RequiredLevel,
		"ClassId":            c.ClassId,
		"SubClassId":         c.SubClassId,
		"IconFileDataID":     c.IconFileDataID,
		"SpellCategoryID":    c.SpellCategoryID,
		"SpellCategoryFlags": c.SpellCategoryFlags,
		"ItemEffects":        c.ItemEffects,
	}
}

// ToProto converts the Consumable to a proto representation.
func (c *Consumable) ToProto() *proto.Consumable {
	consumable := &proto.Consumable{
		Id:                       int32(c.Id),
		Type:                     c.GetConsumableType(),
		Stats:                    c.GetStatModifiers().ToProtoArray(),
		Name:                     c.Name,
		BuffsMainStat:            false, // Todo: Should be food currently, might be more in MoP, figure out how to tell
		BuffDuration:             int32(c.Duration / 1000),
		CooldownDuration:         int32(c.CooldownDuration / 1000),
		CategoryCooldownDuration: int32(c.CategoryCooldownDuration / 1000),
		CategoryId:               int32(c.SpellCategoryID),
		EffectIds:                c.GetNonStatEffectIds(),
		ItemEffects:              c.OnUseEffects(),
		ClassAllowlist:           GetClassesFromClassMask(c.ClassMask),
		RequiredProfession:       GetProfession(c.RequiredSkill),
	}
	if c.RaceMask != -1 {
		consumable.AllowableRace = c.RaceMask
	}
	if enchant := c.temporaryEnchantSpell(); enchant != nil && enchant.EquippedItemClass == ITEM_CLASS_WEAPON {
		consumable.WeaponSubclassMask = enchant.EquippedItemSubclass
		consumable.WeaponInventoryTypeMask = enchant.EquippedItemInvTypes
	}
	return consumable
}

// The spell each on-use row casts, with the cooldowns and category the item states for it.
func (c *Consumable) OnUseEffects() []*proto.ItemEffect {
	var effects []*proto.ItemEffect
	for _, id := range slices.Sorted(slices.Values(c.ItemEffects)) {
		e := GetItemEffect(id)
		if e.ID == 0 || e.TriggerType != ITEM_SPELLTRIGGER_ON_USE || e.SpellID == 0 {
			continue
		}
		effects = append(effects, &proto.ItemEffect{
			BuffId: int32(e.SpellID),
			Effect: &proto.ItemEffect_OnUse{OnUse: &proto.OnUseEffect{
				CooldownMs:         int32(e.CoolDownMSec),
				CategoryId:         int32(e.SpellCategoryID),
				CategoryCooldownMs: int32(e.CategoryCoolDownMSec),
			}},
		})
	}
	return effects
}

func (c *Consumable) temporaryEnchantSpell() *Spell {
	for _, e := range c.OnUseEffects() {
		for _, se := range dbcInstance.SpellEffects[int(e.BuffId)] {
			if se.EffectType == dbcenums.E_ENCHANT_ITEM_TEMPORARY {
				spell := dbcInstance.Spells[int(e.BuffId)]
				return &spell
			}
		}
	}
	return nil
}

func (c *Consumable) sharesExplosivesCategory() bool {
	return slices.ContainsFunc(c.OnUseEffects(), func(e *proto.ItemEffect) bool {
		return e.GetOnUse().CategoryId == EXPLOSIVES_SPELL_CATEGORY
	})
}

func (c *Consumable) GetConsumableType() proto.ConsumableType {
	switch {
	case c.temporaryEnchantSpell() != nil:
		return proto.ConsumableType_ConsumableTypeImbue
	case c.sharesExplosivesCategory(), c.ClassId == ITEM_CLASS_TRADE_GOODS && c.SubClassId == TRADE_GOODS_EXPLOSIVES:
		return proto.ConsumableType_ConsumableTypeExplosive
	case c.ClassId != ITEM_CLASS_CONSUMABLE:
		return proto.ConsumableType_ConsumableTypeUnknown
	}
	if c.SubClassId == ELIXIR {
		switch c.ElixirType {
		case 1:
			return proto.ConsumableType_ConsumableTypeGuardianElixir
		case 2:
			return proto.ConsumableType_ConsumableTypeBattleElixir
		}
	}
	if c.SubClassId == FOOD {
		// Ugly way to check for pet food
		for _, effectID := range c.ItemEffects {
			effect := GetItemEffect(effectID)
			if effect.ID != 0 {
				if spellEffects, ok := dbcInstance.SpellEffects[effect.SpellID]; ok {
					for _, spellEffect := range spellEffects {
						if slices.Contains(spellEffect.ImplicitTargets, 5) {
							return proto.ConsumableType_ConsumableTypePetFood
						}
					}
				}
			}
		}

		return proto.ConsumableType_ConsumableTypeFood
	}
	if val, ok := consumableClassToProto[c.SubClassId]; ok {
		return val
	}
	return proto.ConsumableType_ConsumableTypeUnknown
}

func (s ConsumableClass) ToProto() proto.ConsumableType {
	if val, ok := consumableClassToProto[s]; ok {
		return val
	}
	return proto.ConsumableType_ConsumableTypeUnknown
}

// A consumable's effects that restore a resource rather than grant a stat, which the sim wires up
// separately.
var consumableRestoreEffectTypes = map[SpellEffectType]bool{
	dbcenums.E_HEAL:     true,
	dbcenums.E_ENERGIZE: true,
}

func (consumable *Consumable) GetNonStatEffectIds() []int32 {
	var effectIds []int32

	slices.Sort(consumable.ItemEffects)
	for _, effectID := range consumable.ItemEffects {
		effect := GetItemEffect(effectID)
		if effect.ID != 0 {
			if spellEffects, ok := dbcInstance.SpellEffects[effect.SpellID]; ok {
				for _, spellEffect := range spellEffects {
					if consumableRestoreEffectTypes[spellEffect.EffectType] || spellEffect.EffectAura == dbcenums.A_PERIODIC_ENERGIZE || spellEffect.EffectAura == dbcenums.A_PERIODIC_HEAL {
						effectIds = append(effectIds, int32(spellEffect.ID))
					}
					// Aura effects triggered on use (e.g. Fel Mana Potion's spell damage
					// reduction) carry the stats the sim applies as a temporary aura.
					if spellEffect.EffectType == dbcenums.E_TRIGGER_SPELL {
						for _, subEffect := range dbcInstance.SpellEffects[spellEffect.EffectTriggerSpell] {
							if _, ok := subEffect.ParseStatEffect(false, 0); ok {
								effectIds = append(effectIds, int32(subEffect.ID))
							}
						}
					}
				}
			}
		}
	}
	slices.Sort(effectIds)
	return effectIds
}
func (consumable *Consumable) GetStatModifiers() *stats.Stats {
	stats := &stats.Stats{}
	for _, effectID := range consumable.ItemEffects {
		effect := GetItemEffect(effectID)
		if effect.ID != 0 {
			if spellEffects, ok := dbcInstance.SpellEffects[effect.SpellID]; ok {
				for _, spellEffect := range spellEffects {
					if stat, ok := spellEffect.ParseStatEffect(spellEffect.Coefficient != 0, 0); ok {
						stats.AddInplace(&stat)
					}
				}
			}
		}
	}
	return stats
}
