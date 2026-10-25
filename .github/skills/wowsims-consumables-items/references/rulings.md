# Rulings

Each line is the rule the code follows, where it lives, and the pull request that settled it.

## Consumables

- **Caustic Smog Potion 250951 and Dragonfire Potion 250953 are category 4 offensive potions, never explosives.
  Caustic Smog's missile flies at its Speed for no less than its minimum flight time, and its cloud puts the damage
  over time on every enemy unrolled; Dragonfire's Fixed Travel Time makes its Speed a flight time in seconds, and it
  cannot be used past its cone's radius.** `sim/common/shared/consumable_explosives.go`, `cloudDamageSpell`, PR #573.
- **Elixirs stack their stats across groups, but only the strongest elixir crit applies.**
  `sim/common/shared/consumable_stats.go`, `consumableAura` with `spelldata.Exclusive(elixirCritCategory, false)` (the category `"ElixirCrit"`), PR #209.
- **Holiday consumables (`Consumable.holiday`) sit in their own row above Miscellaneous, are not Well Fed and stack
  with food; a Forever food's stats are the amount Nutritious Food's `A_PERIODIC_TRIGGER_SPELL_WITH_VALUE` hands Well
  Fed.** `tools/database/gen_consumables.go`, `consumableHoliday`; `sim/common/shared/consumable_stats.go`,
  `ConsumableSpells`, PR #207.
- **`Consumable.mob_type_stats` and `zone_stats` gate a consumable's EP by the target's mob type and the encounter
  zone, and consumables are ordered by EP.** `ui/sim/settings/consumables.ts`, `orderedByEp`, PR #344.
- **The swap set carries its own stones, oils and weapon imbues (`ConsumesSpec` swap fields 24 to 27), applied
  through one `applyTemporaryEnchants` against the given item, and `Item.TempEnchantStats` counts as equipment stats,
  not consumable stats.** `sim/core/consumes_imbues.go`, PR #338.

## Item effects

- **A hand-written item effect only one class can use lives in `sim/<class>/item_<kind>.go` (sets, librams, idols,
  totems, trinkets, weapons); class-agnostic items live in `sim/common/forever/items_*.go`. A set the game files open
  to every class is shared, in `sim/common/forever/items_sets.go` with `agent.GetCharacter()`, and so are the dungeon
  sets; Naxxramas tier 3 armor is class-locked through `ItemSetClasses` and coded in the class folders.**
  `tools/database/overrides.go`, `ApplyItemSetClasses`, PR #316 and PR #539.
- **Equip spell modifiers are built by hand per class with `shared.NewSpellDataEquipAura` in
  `sim/<class>/item_*.go`; unported classes register their item effects anyway; items whose modifier has no sim value
  stay in the missing list.** `sim/common/shared/spell_data_aura.go`, PR #322.
- **Survival cooldowns default to `ShouldActivate` false unless the character sits in a raid Tanks slot; defensive
  on-use buffs count as Survival and MP5 or Spirit ones as Mana; explicit APL casts are unaffected.**
  `sim/core/major_cooldown.go`, PR #335.
- **Creature-type multipliers apply against the target's mob type; party auras (Atiesh, Darkspear, Admiral's Hat)
  are a wearer effect plus a party buff toggle; the row wins over the tooltip (Ferra's Collar +2 flat, Aegis of the
  Scarlet Commander's 5 s RecoveryTime as its ICD).** `sim/common/forever/items_party.go`,
  `sim/mage/item_weapons.go`, PR #322.
- **Ravager's 9 s Bladestorm channel locks the wearer and its strikes can proc other weapon effects but not Ravager;
  Hand of Edward's focus is spent only by cast-time spells, which become instant while the GCD stays 1.5 s.**
  `sim/common/forever/items_weapons.go`, PR #322.

## Item data

- **ItemSparse `REDUCED_CASTER_DPS` gives weapon DPS = damage table × 2/3 one-hand or × 29/39 two-hand, with
  QualityModifier an additive offset; the shares are fitted to in-game ranges, not read from the game files.**
  `tools/database/dbc/item_damage.go`, `reducedCasterDpsShare`, PR #298.
- **`CASTER_SPELL_POWER` gives spell power ⌊20000 × budget / 10000⌋ and `CASTER_HEALING_POWER` gives healing
  ⌊37600 × budget / 10000⌋ with spell damage ⌊healing / 3⌋, budget being RandPropPoints column 0.**
  `tools/database/dbc/item.go`, `addCasterSpellPower`, PR #302.
- **Presets carry no default random suffix, the UI adds suffix stats to the item's, and flat dodge goes through
  `ItemRandomSuffix.pseudo_stats`.** `ui/sim/proto/equipped_item.ts`, `tools/database/dbc/random_suffix.go`,
  PR #303.
- **Gems, sockets, socket bonuses and Jewelcrafting are removed; no gem code comes back.** PR #216.
