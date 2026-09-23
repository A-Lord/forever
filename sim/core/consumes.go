package core

import (
	"fmt"
	"slices"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Scrolls share StatBuffCategory with the raid buffs granting the same stat, so a
// scroll doesn't stack with e.g. Arcane Brilliance or Divine Spirit; only the
// strongest source of that stat applies.
func registerScrollAura(character *Character, label string, itemID int32, stat stats.Stat, amount float64) *Aura {
	aura := character.GetOrRegisterAura(Aura{
		Label:      label,
		ActionID:   ActionID{ItemID: itemID},
		Duration:   NeverExpires,
		BuildPhase: CharacterBuildPhaseConsumes,
	})
	makeExclusiveFlatStatBuff(aura, stat, amount, StatBuffCategory)
	return MakePermanent(aura)
}

// Registers all consume-related effects to the Agent.
func applyConsumeEffects(agent Agent, partyBuffs *proto.PartyBuffs) {
	character := agent.GetCharacter()
	consumables := character.Consumables
	if consumables == nil {
		return
	}

	if consumables.FlaskId != 0 {
		flask := GetConsumableByID(consumables.FlaskId)
		character.AddStats(flask.Stats)
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
		} else {
			elixir := GetConsumableByID(consumables.BattleElixirId)
			character.AddStats(elixir.Stats)
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
		} else {
			elixir := GetConsumableByID(consumables.GuardianElixirId)
			character.AddStats(elixir.Stats)
		}
	}
	if consumables.FoodId != 0 {
		food := GetConsumableByID(consumables.FoodId)
		character.AddStats(food.Stats)
	}

	applyTemporaryEnchants(character, consumables, partyBuffs)

	// Scrolls
	if consumables.ScrollAgi {
		registerScrollAura(character, "Scroll of Agility", 27498, stats.Agility, 20)
	}
	if consumables.ScrollStr {
		registerScrollAura(character, "Scroll of Strength", 27503, stats.Strength, 20)
	}
	if consumables.ScrollInt {
		registerScrollAura(character, "Scroll of Intellect", 27499, stats.Intellect, 20)
	}
	if consumables.ScrollSpi {
		registerScrollAura(character, "Scroll of Spirit", 27501, stats.Spirit, 30)
	}
	if consumables.ScrollArm {
		registerScrollAura(character, "Scroll of Protection", 27500, stats.Armor, 300)
	}

	// Bloodthistle (Blood Elf only): +10 spell damage and healing for 10 min.
	if consumables.Bloodthistle && character.Race == proto.Race_RaceBloodElf {
		character.AddStat(stats.SpellDamage, 10)
		character.AddStat(stats.HealingPower, 10)
	}

	// Bogling Root: +1 physical damage for 10 min (item 5206, spell 5665).
	if consumables.BoglingRoot {
		character.AddStat(stats.PhysicalDamage, 1)
	}

	// Pet Consumes
	for _, pet := range character.Pets {
		if pet.isGuardian {
			continue
		}

		if consumables.PetScrollAgi {
			pet.AddStat(stats.Agility, 20)
		}
		if consumables.PetScrollStr {
			pet.AddStat(stats.Strength, 20)
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

// Goblin Sapper Charge, which the spec selects with a flag of its own.
const goblinSapperItemID = 10646

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
// Windfury Totem, as the hand-written imbues are; the pet's food is the pet's.
func selectedConsumableIDs(consumables *proto.ConsumesSpec, partyBuffs *proto.PartyBuffs) []int32 {
	ids := []int32{consumables.PotId, consumables.FlaskId, consumables.BattleElixirId, consumables.GuardianElixirId,
		consumables.FoodId, consumables.ConjuredId, consumables.ExplosiveId, consumables.OhImbueId}
	if partyBuffs.WindfuryTotem == proto.TristateEffect_TristateEffectMissing {
		ids = append(ids, consumables.MhImbueId)
	}
	if consumables.GoblinSapper {
		ids = append(ids, goblinSapperItemID)
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
