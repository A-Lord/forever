package database

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/tools/database/dbc"
)

// The shape a consumable's on-use spell is built in. Damage and heals are routed the way an item's
// on-use is, and have builders; the other shapes name the builder a consumable of that kind needs.
type ConsumableKind string

const (
	ConsumableDamage           ConsumableKind = "damage"
	ConsumableHeal             ConsumableKind = "heal"
	ConsumableEnergize         ConsumableKind = "energize"
	ConsumableStatAura         ConsumableKind = "stat aura"
	ConsumablePercentStat      ConsumableKind = "percent stat"
	ConsumableTemporaryEnchant ConsumableKind = "temporary enchant"
	ConsumableAbsorb           ConsumableKind = "absorb"
)

var ConsumableKinds = []ConsumableKind{
	ConsumableDamage, ConsumableHeal, ConsumableEnergize, ConsumableStatAura, ConsumablePercentStat,
	ConsumableTemporaryEnchant, ConsumableAbsorb,
}

type ConsumableRouting struct {
	ItemID  int
	SpellID int32
	// Empty when no shape fits the spell.
	Kind        ConsumableKind
	Unsupported []string
}

func (r *ConsumableRouting) Supported() bool {
	return r.Kind != "" && len(r.Unsupported) == 0
}

func (r *ConsumableRouting) Reason() string {
	return strings.Join(r.Unsupported, "; ")
}

// Whether the catalogue's consumable goes into db.json: its spell routes to a shape, or
// ConsumableAllowList or ConsumableOverrides names it. KeepConsumable runs after.
func ShipsConsumable(routing *ConsumableRouting) bool {
	id := int32(routing.ItemID)
	return routing.Supported() || slices.Contains(ConsumableAllowList, id) ||
		slices.ContainsFunc(ConsumableOverrides, func(c *proto.Consumable) bool { return c.Id == id })
}

// Whether a consumable in db.json survives the global filters. ConsumableAllowList keeps it,
// ConsumableDenyList drops it, and so does a test or QA name.
func KeepConsumable(consumable *proto.Consumable) bool {
	if slices.Contains(ConsumableAllowList, consumable.Id) {
		return true
	}
	if slices.Contains(ConsumableDenyList, consumable.Id) {
		return false
	}
	for _, pattern := range DenyListNameRegexes {
		if pattern.MatchString(consumable.Name) {
			return false
		}
	}

	return !strings.HasPrefix(consumable.Name, "QA") && !strings.HasPrefix(consumable.Name, "Test") && !strings.HasPrefix(consumable.Name, "TEST") && !strings.Contains(consumable.Name, "Flaskataur")
}

// Routes the spell the consumable's first on-use row casts: as damage, a heal or an absorb where the
// on-use routing takes it, and otherwise by the first of its effects, or of the spells it triggers on
// the user, that fits a consumable shape. A heal, an absorb or a gain of a resource is supported only
// where the potion builder runs the spell.
func RouteConsumable(c *dbc.Consumable, instance *dbc.DBC) *ConsumableRouting {
	routing := &ConsumableRouting{ItemID: c.Id}
	onUses := c.OnUseEffects()
	if len(onUses) == 0 {
		routing.Unsupported = []string{"the item states no on-use spell"}
		return routing
	}
	routing.SpellID = onUses[0].BuffId

	onUse := routeOnUseSpell(c.Id, routing.SpellID, instance)
	switch {
	case onUse.Supported() && onUse.Damage:
		routing.Kind = ConsumableDamage
		routing.Unsupported = damageConsumableUnsupported(spelldata.Find(routing.SpellID))
		return routing
	case onUse.Supported() && onUse.Heal:
		routing.Kind = ConsumableHeal
		routing.Unsupported = shared.ConsumableUnsupported(spelldata.Find(routing.SpellID))
		return routing
	case onUse.Absorb:
		routing.Kind = ConsumableAbsorb
		routing.Unsupported = append(onUse.Unsupported, shared.ConsumableUnsupported(spelldata.Find(routing.SpellID))...)
		return routing
	}

	routing.Kind, routing.Unsupported = consumableShape(c, spelldata.Find(routing.SpellID))
	if routing.Kind == "" {
		routing.Unsupported = onUse.Unsupported
	}
	return routing
}

func consumableShape(c *dbc.Consumable, s *spelldata.Spell) (ConsumableKind, []string) {
	var seated *spelldata.Effect
	for _, e := range consumableEffects(s) {
		switch {
		case dbc.IsTemporaryEnchantEffect(e.Type):
			return ConsumableTemporaryEnchant, temporaryEnchantUnsupported(c.TemporaryEnchant)
		case !landsOnTheUser(e):
		case e.Type == dbcenums.E_ENERGIZE, e.Type == dbcenums.E_ENERGIZE_PCT,
			e.Aura == dbcenums.A_PERIODIC_ENERGIZE, e.Aura == dbcenums.A_OBS_MOD_POWER:
			return ConsumableEnergize, shared.ConsumableUnsupported(s)
		case e.Aura == dbcenums.A_SCHOOL_ABSORB:
			return ConsumableAbsorb, []string{fmt.Sprintf("%d absorbs damage, which no builder takes from the rows", e.SpellID)}
		case e.Aura == dbcenums.A_MOD_PERCENT_STAT, e.Aura == dbcenums.A_MOD_TOTAL_STAT_PERCENTAGE:
			return ConsumablePercentStat, auraDurationUnsupported(e)
		case e.ParsedAura() && spelldata.Find(e.SpellID).WhileSeated():
			seated = cmp.Or(seated, e)
		case e.ParsedAura():
			return ConsumableStatAura, statAuraUnsupported(s)
		}
	}
	if seated != nil {
		return ConsumableStatAura, append([]string{fmt.Sprintf("%d restores only while the user sits, out of combat (%v)",
			seated.SpellID, seated.Aura)}, statAuraUnsupported(s)...)
	}
	return "", nil
}

// The auras of a stat consumable's spells that the stat builder cannot put up: one the parser does
// not attach, other than the model's size, one whose amount the server rolls, and a spell that states
// no duration for its aura. The eating and drinking are left out, since they last only while seated.
func statAuraUnsupported(s *spelldata.Spell) []string {
	var unsupported []string
	for _, e := range consumableEffects(s) {
		spell := spelldata.Find(e.SpellID)
		switch {
		case !landsOnTheUser(e), e.Aura == 0, spell.WhileSeated():
		case e.Aura == dbcenums.A_PERIODIC_TRIGGER_SPELL, e.Aura == dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE,
			e.Aura == dbcenums.A_MOD_SCALE:
		case !e.ParsedAura():
			unsupported = append(unsupported, fmt.Sprintf("%d applies %v, which the parser does not attach", e.SpellID, e.Aura))
		case e.Variance != 0:
			unsupported = append(unsupported, fmt.Sprintf("%d rolls its %v amount (variance %v), which the parser reads as its average",
				e.SpellID, e.Aura, e.Variance))
		default:
			unsupported = append(unsupported, auraDurationUnsupported(e)...)
		}
	}
	return slices.Compact(unsupported)
}

// The user, or the one friendly unit the spell is cast on, which in the sim is the user: a scroll
// states the latter.
func landsOnTheUser(e *spelldata.Effect) bool {
	return slices.Contains([]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER, dbcenums.TARGET_UNIT_TARGET_ALLY,
		dbcenums.TARGET_UNIT_TARGET_ANY}, e.Target[0])
}

// The spell's effects, followed by those of each spell it triggers on the user: food puts its
// Well Fed buff up that way.
func consumableEffects(s *spelldata.Spell) []*spelldata.Effect {
	var effects, triggered []*spelldata.Effect
	for i := range s.Effects {
		e := &s.Effects[i]
		effects = append(effects, e)
		if e.TriggerID == 0 || !landsOnTheUser(e) || (e.Type != dbcenums.E_TRIGGER_SPELL &&
			e.Aura != dbcenums.A_PERIODIC_TRIGGER_SPELL && e.Aura != dbcenums.A_PERIODIC_TRIGGER_SPELL_WITH_VALUE) {
			continue
		}
		trigger := e.Trigger()
		for j := range trigger.Effects {
			triggered = append(triggered, &trigger.Effects[j])
		}
	}
	return append(effects, triggered...)
}

func auraDurationUnsupported(e *spelldata.Effect) []string {
	if spelldata.Find(e.SpellID).DurationMs <= 0 {
		return []string{fmt.Sprintf("%d states no duration for its aura", e.SpellID)}
	}
	return nil
}

// A temporary enchant's slots, by what the sim's enchant handling reads: flat weapon damage, a
// resistance or a stat, a combat spell that deals damage at the chance the slot states, or an equip
// spell whose auras the parser attaches. Spells that work only in one area do nothing in the sim.
func temporaryEnchantUnsupported(enchant dbc.TemporaryEnchant) []string {
	if enchant.ID == 0 {
		return []string{"the temporary enchant names no SpellItemEnchantment row"}
	}

	var unsupported []string
	for slot, effect := range enchant.Effects {
		spellID := int32(enchant.Args[slot])
		switch effect {
		case dbc.ITEM_ENCHANTMENT_NONE, dbc.ITEM_ENCHANTMENT_DAMAGE, dbc.ITEM_ENCHANTMENT_RESISTANCE, dbc.ITEM_ENCHANTMENT_STAT:
		case dbc.ITEM_ENCHANTMENT_COMBAT_SPELL:
			if !dealsDamageToAnEnemy(spelldata.Find(spellID)) {
				unsupported = append(unsupported, fmt.Sprintf("enchant %d casts %d, which deals no damage to an enemy", enchant.ID, spellID))
			}
			if enchant.Points[slot] <= 0 {
				unsupported = append(unsupported, fmt.Sprintf("enchant %d states no chance for its combat spell %d", enchant.ID, spellID))
			}
			if s := spelldata.Find(spellID); s.PeriodicDamageEffect() != spelldata.NilEffect && s.MaxStack > 0 {
				unsupported = append(unsupported, fmt.Sprintf("enchant %d casts %d, whose damage over time stacks to %d, which the proc builder applies as one", enchant.ID, spellID, s.MaxStack))
			}
		case dbc.ITEM_ENCHANTMENT_EQUIP_SPELL:
			unsupported = append(unsupported, equipSpellUnsupported(enchant.ID, spelldata.Find(spellID), spellID)...)
		default:
			unsupported = append(unsupported, fmt.Sprintf("enchant %d has a slot of type %d", enchant.ID, effect))
		}
	}
	if enchant.RequiredAreasID != 0 {
		unsupported = append(unsupported, fmt.Sprintf("enchant %d's spells work only in area group %d", enchant.ID, enchant.RequiredAreasID))
	}
	return unsupported
}

func dealsDamageToAnEnemy(s *spelldata.Spell) bool {
	for _, e := range []*spelldata.Effect{s.DamageEffect(), s.PeriodicDamageEffect()} {
		if e != spelldata.NilEffect && e.HitsAnEnemy() {
			return true
		}
	}
	return false
}

func equipSpellUnsupported(enchantID int, s *spelldata.Spell, spellID int32) []string {
	if s == spelldata.Nil {
		return []string{fmt.Sprintf("enchant %d's equip spell %d has no row in the store", enchantID, spellID)}
	}
	var unsupported []string
	for i := range s.Effects {
		if e := &s.Effects[i]; e.Aura != 0 && !e.ParsedAura() {
			unsupported = append(unsupported, fmt.Sprintf("enchant %d's equip spell %d applies %v, which the parser does not attach", enchantID, spellID, e.Aura))
		}
	}
	return unsupported
}
