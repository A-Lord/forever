package database

import (
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/tools/database/dbc"
)

// An exclusive group the rows do not state. Items join the row group Joins names, or with an empty
// Joins form a group of their own, keyed and labelled by Name.
type ConsumableGroup struct {
	Name  string
	Joins ConsumableGroupKey
	Items []int32
}

type ConsumableGroupKey struct {
	Key  string
	Name string
}

// The group of the elixirs whose on-use spell's first aura is the aura type with the misc value.
func ElixirGroup(aura dbcenums.EffectAuraType, misc int) ConsumableGroupKey {
	return auraGroup("elixir", aura, misc)
}

func scrollGroup(aura dbcenums.EffectAuraType, misc int) ConsumableGroupKey {
	return auraGroup("scroll", aura, misc)
}

func auraGroup(kind string, aura dbcenums.EffectAuraType, misc int) ConsumableGroupKey {
	return ConsumableGroupKey{Key: fmt.Sprintf("%s %d/%d", kind, aura, misc), Name: auraGroupName(aura, misc)}
}

// Writes the consumable's exclusive group. ConsumableExclusiveGroups names it first. Otherwise an
// elixir's is the aura type and misc value of the first aura its on-use spell applies, and a scroll's
// the same among the scrolls, so the ranks of one scroll share a group and a scroll stacks with every
// elixir. A stat consumable used one at a time with the others of its category
// (shared.OneAtATimeCategory) is grouped by that category. Any other consumable has no group.
func setConsumableGroup(c *dbc.Consumable, consumable *proto.Consumable, routing *ConsumableRouting, instance *dbc.DBC) {
	group := consumableGroupOf(c, consumable, routing, instance)
	consumable.GroupKey, consumable.GroupName = group.Key, group.Name
}

func consumableGroupOf(c *dbc.Consumable, consumable *proto.Consumable, routing *ConsumableRouting, instance *dbc.DBC) ConsumableGroupKey {
	if group, ok := exclusiveGroupOverride(consumable.Id); ok {
		return group
	}
	return rowGroupOf(c, consumable, routing, instance)
}

func rowGroupOf(c *dbc.Consumable, consumable *proto.Consumable, routing *ConsumableRouting, instance *dbc.DBC) ConsumableGroupKey {
	if len(consumable.ItemEffects) == 0 {
		return ConsumableGroupKey{}
	}
	onUse := consumable.ItemEffects[0]
	switch consumable.Type {
	case proto.ConsumableType_ConsumableTypeElixir, proto.ConsumableType_ConsumableTypeScroll:
		aura, misc, ok := firstAura(instance, int(onUse.BuffId))
		if !ok {
			return ConsumableGroupKey{}
		}
		if consumable.Type == proto.ConsumableType_ConsumableTypeScroll {
			return scrollGroup(aura, misc)
		}
		return ElixirGroup(aura, misc)
	}
	if runsAsStatConsumable(c, consumable, routing) {
		if category := shared.OneAtATimeCategory(onUse); category > 0 {
			return ConsumableGroupKey{Key: fmt.Sprintf("category %d", category), Name: fmt.Sprintf("Category %d", category)}
		}
	}
	return ConsumableGroupKey{}
}

func exclusiveGroupOverride(id int32) (ConsumableGroupKey, bool) {
	for _, group := range ConsumableExclusiveGroups {
		if !slices.Contains(group.Items, id) {
			continue
		}
		if group.Joins.Key != "" {
			return group.Joins, true
		}
		return ConsumableGroupKey{Key: group.Name, Name: group.Name}, true
	}
	return ConsumableGroupKey{}, false
}

// Whether NewSpellDataStatConsumable registers the consumable, as GenerateConsumableRegistrations
// picks it.
func runsAsStatConsumable(c *dbc.Consumable, consumable *proto.Consumable, routing *ConsumableRouting) bool {
	switch {
	case consumable.Type == proto.ConsumableType_ConsumableTypeImbue, IsExplosive(c, routing),
		consumable.CategoryId == dbc.POTION_SPELL_CATEGORY, consumable.Type == proto.ConsumableType_ConsumableTypeConjured:
		return false
	}
	return routing.Supported() && routing.Kind == ConsumableStatAura
}

func firstAura(instance *dbc.DBC, spellID int) (dbcenums.EffectAuraType, int, bool) {
	for _, e := range instance.SpellEffectsInOrder(spellID) {
		if e.EffectType == dbcenums.E_APPLY_AURA {
			misc := 0
			if len(e.EffectMiscValues) > 0 {
				misc = e.EffectMiscValues[0]
			}
			return e.EffectAura, misc, true
		}
	}
	return 0, 0, false
}

var modStatNames = []string{"Strength", "Agility", "Stamina", "Intellect", "Spirit"}

const allSpellSchools = 126

func auraGroupName(aura dbcenums.EffectAuraType, misc int) string {
	switch aura {
	case dbcenums.A_MOD_STAT:
		if misc >= 0 && misc < len(modStatNames) {
			return modStatNames[misc]
		}
		return "All Stats"
	case dbcenums.A_MOD_DAMAGE_DONE:
		switch misc {
		case allSpellSchools:
			return "Spell Power"
		case int(core.SpellSchoolPhysical):
			return "Physical Damage"
		}
		return schoolLabel(misc) + " Spell Power"
	case dbcenums.A_MOD_RESISTANCE:
		if misc == int(core.SpellSchoolPhysical) {
			return "Armor"
		}
		return schoolLabel(misc) + " Resistance"
	case dbcenums.A_MOD_INCREASE_HEALTH:
		return "Health"
	case dbcenums.A_MOD_POWER_REGEN, dbcenums.A_PERIODIC_ENERGIZE:
		return "Mana Regeneration"
	case dbcenums.A_MOD_ATTACK_POWER:
		return "Attack Power"
	case dbcenums.A_MOD_HEALING_DONE:
		return "Healing"
	case dbcenums.A_MOD_CRIT_PCT:
		return "Critical Strike"
	}
	return "Other"
}

func schoolLabel(mask int) string {
	name := core.SpellSchool(mask).String()
	if strings.HasPrefix(name, "SpellSchool(") {
		return "School"
	}
	return strings.TrimPrefix(name, "SpellSchool")
}
