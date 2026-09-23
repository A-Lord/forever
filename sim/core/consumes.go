package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A consumable the spec selects with a flag rather than an id. The pet's flags are the pet's.
type ConsumableFlag struct {
	Field  protoreflect.Name
	ItemID int32
	Pet    bool
}

// Goblin Sapper Charge, the highest rank of each scroll the client ships, and Bogling Root, by the
// ConsumesSpec field that selects them. A request's database carries these items for its flags.
var ConsumableFlags = []ConsumableFlag{
	{"goblinSapper", 10646, false},
	{"scroll_agi", 10309, false}, // Scroll of Agility IV
	{"scroll_str", 10310, false}, // Scroll of Strength IV
	{"scroll_int", 10308, false}, // Scroll of Intellect IV
	{"scroll_spi", 10306, false}, // Scroll of Spirit IV
	{"scroll_arm", 10305, false}, // Scroll of Protection IV
	{"bogling_root", 5206, false},
	{"pet_scroll_agi", 10309, true},
	{"pet_scroll_str", 10310, true},
}

func (flag ConsumableFlag) IsSet(consumables *proto.ConsumesSpec) bool {
	message := consumables.ProtoReflect()
	return message.Get(message.Descriptor().Fields().ByName(flag.Field)).Bool()
}

// Registers all consume-related effects to the Agent.
func applyConsumeEffects(agent Agent, partyBuffs *proto.PartyBuffs) {
	character := agent.GetCharacter()
	consumables := character.Consumables
	if consumables == nil {
		return
	}

	if consumables.BattleElixirId != 0 {
		// Potion of Demonslaying, spell 11406: melee attack power and spell damage against demons.
		if consumables.BattleElixirId == 9224 {
			character.Env.RegisterPostFinalizeEffect(func() {
				for _, at := range character.AttackTables {
					at.MobTypeBonusStats[proto.MobType_MobTypeDemon] = at.MobTypeBonusStats[proto.MobType_MobTypeDemon].Add(stats.Stats{
						stats.AttackPower: 185,
						stats.SpellDamage: 108,
					})
				}
			})
		}
	}

	if consumables.GuardianElixirId != 0 {
		// Gift of Arthas
		if consumables.GuardianElixirId == 9088 {
			character.AddStat(stats.ShadowResistance, 10)
			auras := character.NewEnemyAuraArray(func(target *Unit) *Aura {
				return GiftOfArthasAura(target)
			})
			procSpell := character.RegisterSpell(SpellConfig{
				ActionID:    ActionID{SpellID: 11374},
				SpellSchool: SpellSchoolNature,
				ProcMask:    ProcMaskEmpty,

				FlatThreatBonus: 90,

				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
					auras.Get(target).Activate(sim)
				},
			})

			character.MakeProcTriggerAura(ProcTrigger{
				Name:       "Gift of Arthas - Trigger",
				ICD:        time.Second * 3,
				ProcChance: 0.3,
				Outcome:    OutcomeLanded,
				Callback:   CallbackOnSpellHitTaken,
				Handler: func(sim *Simulation, spell *Spell, _ *SpellResult) {
					procSpell.Cast(sim, spell.Unit)
				},
			})
		}
	}

	applyTemporaryEnchants(character, consumables, partyBuffs)

	// Pet Consumes
	for _, pet := range character.Pets {
		if pet.isGuardian {
			continue
		}

		for _, flag := range ConsumableFlags {
			if flag.Pet && flag.IsSet(consumables) {
				pet.AddStats(GetConsumableByID(flag.ItemID).Stats)
			}
		}
		if consumables.PetFoodId != 0 {
			petFood := GetConsumableByID(consumables.PetFoodId)
			pet.AddStats(petFood.Stats)
		}
	}

	drumsBombsSharedTimer := character.NewTimer()

	registerDrumsCD(agent, consumables, drumsBombsSharedTimer)

	for _, id := range selectedConsumableIDs(consumables, partyBuffs) {
		if apply, ok := consumableEffect(id); ok {
			apply(agent)
		}
	}
}

var consumableEffects = map[int32]ApplyEffect{}

// The effects built from the consumables' rows. A hand-written effect for the same id wins.
var dataDrivenConsumableEffects = map[int32]ApplyEffect{}

var missingConsumableEffects []int32

func consumableEffect(id int32) (ApplyEffect, bool) {
	if apply, ok := consumableEffects[id]; ok {
		return apply, true
	}
	apply, ok := dataDrivenConsumableEffects[id]
	return apply, ok
}

func HasConsumableEffect(id int32) bool {
	_, ok := consumableEffect(id)
	return ok
}

// MissingConsumableEffects reports the consumable ids whose effects were skipped because the loaded
// database has no such consumable.
func MissingConsumableEffects() []int32 { return missingConsumableEffects }

// Registers a hand-written ApplyEffect function which will be called while the consumes are applied,
// for any Agent whose ConsumesSpec names the id. It takes the place of a data-driven registration for
// the id, whichever of the two registers first.
func NewConsumableEffect(id int32, consumableEffect ApplyEffect) {
	if !consumableInDatabase(id) {
		return
	}

	if _, ok := consumableEffects[id]; ok {
		panic(fmt.Sprintf("Cannot add multiple effects for one consumable: %d, %#v", id, consumableEffect))
	}

	consumableEffects[id] = consumableEffect
}

// Registers an ApplyEffect function built from the consumable's rows, which applies unless a
// hand-written one is registered for the id.
func NewDataDrivenConsumableEffect(id int32, consumableEffect ApplyEffect) {
	if !consumableInDatabase(id) {
		return
	}

	if _, ok := dataDrivenConsumableEffects[id]; ok {
		panic(fmt.Sprintf("Cannot add multiple data-driven effects for one consumable: %d, %#v", id, consumableEffect))
	}

	dataDrivenConsumableEffects[id] = consumableEffect
}

func consumableInDatabase(id int32) bool {
	if WITH_DB && GetConsumableByID(id).Id == 0 {
		missingConsumableEffects = append(missingConsumableEffects, id)
		return false
	}
	return true
}

// The agent's own consumables the spec names, each once. The main-hand imbue is left out under
// Windfury Totem, which applyTemporaryEnchants keeps off the main hand too; the pet's food and
// scrolls are the pet's.
func selectedConsumableIDs(consumables *proto.ConsumesSpec, partyBuffs *proto.PartyBuffs) []int32 {
	ids := []int32{consumables.PotId, consumables.FlaskId, consumables.BattleElixirId, consumables.GuardianElixirId,
		consumables.FoodId, consumables.ConjuredId, consumables.ExplosiveId, consumables.OhImbueId}
	if partyBuffs.WindfuryTotem == proto.TristateEffect_TristateEffectMissing {
		ids = append(ids, consumables.MhImbueId)
	}
	for _, flag := range ConsumableFlags {
		if !flag.Pet && flag.IsSet(consumables) {
			ids = append(ids, flag.ItemID)
		}
	}
	ids = append(ids, consumables.Potions...)
	ids = append(ids, consumables.ConjuredItems...)

	var selected []int32
	for _, id := range ids {
		if id != 0 && !slices.Contains(selected, id) {
			selected = append(selected, id)
		}
	}
	return selected
}
