# Consumables

A consumable reaches the sim in two halves: its entry in `assets/database/db.json`, which the UI offers
and the sim looks the id up in, and an effect registered for its id, which the sim applies when the
character's `ConsumesSpec` names it. The generator writes the first from the client, and writes the
second for every consumable a builder in `sim/common/shared` can run from its rows. Where the rows are
not enough or not right, the effect is hand-written.

- [How a consumable gets into db.json](#how-a-consumable-gets-into-dbjson)
- [Regenerating](#regenerating)
- [How the effect is registered](#how-the-effect-is-registered)
- [Writing an effect by hand: the Natural flasks](#writing-an-effect-by-hand-the-natural-flasks)
- [Adding a consumable the rows do not support](#adding-a-consumable-the-rows-do-not-support)
- [Removing a consumable](#removing-a-consumable)
- [Correcting a number](#correcting-a-number)
- [Tests](#tests)
- [The browser path](#the-browser-path)

## How a consumable gets into db.json

**The catalogue.** `ConsumableItemFilter` in `tools/database/tables.go` selects every item that states
an on-use spell and is one of:

- a Consumable-class item, except the permanent enhancements of subclass 6, which the enchant table
  carries;
- a Trade Goods explosive;
- an item whose on-use shares the explosives' spell category or casts a temporary enchant, whatever
  class the client files it under: Crystal Charge 11566 is a Quest item.

An item that also teaches a spell is an Engrave rune and is left out, and so is a name containing
`Test` or starting with `QA`. The ids in `ConsumableAllowList` (`tools/database/overrides.go`) join the
catalogue by id, whatever the filter says. The catalogue's spells, and everything they trigger, are
carried in the spell store.

**Routing.** `RouteConsumable` in `tools/database/gen_consumables.go` reads the spell the consumable's
first on-use row casts. Damage, heals and absorbs are routed the way an item's on-use is; otherwise the
first effect of the spell, or of a spell it triggers on the user, that fits a shape decides it:

| kind                | what the spell does                                                    |
| ------------------- | ---------------------------------------------------------------------- |
| `damage`            | deals damage (Goblin Sapper Charge 10646: 13241)                       |
| `heal`              | heals the user (Major Healing Potion 13446: 17534)                     |
| `energize`          | restores a resource (Major Mana Potion 13444: 17531)                   |
| `stat aura`         | applies an aura the parser attaches (Flask of the Titans 13510: 17626) |
| `percent stat`      | multiplies a stat; none routes there in this client                    |
| `temporary enchant` | enchants a weapon (Brilliant Wizard Oil 20749: enchant 2628)           |
| `absorb`            | shields the user (Greater Fire Protection Potion 13457: 17543)         |

A consumable that fits no shape, or fits one the rows do not fully state (an aura with no duration, an
enchant slot the sim does not read), is unsupported, with a reason. A heal, an energize or an absorb is
supported only where the potion builder runs the spell: `shared.ConsumableUnsupported` names what stops
it, such as a stun on the drinker (Dreamless Sleep Potion 12190: 15822), an `E_DUMMY` beside the heal
(the Discolored Healing Potions) or a spell that restores only while the user sits (Lollipop 20388:
24707).

**What ships.** `ShipsConsumable` puts a catalogue consumable into db.json when it routes to a shape, or
when `ConsumableAllowList` or `ConsumableOverrides` names it. `ConsumableOverrides` is merged next, field
by field onto the client's entry, or as a new entry where the client has none. `KeepConsumable` then
drops what `ConsumableDenyList` names and the test and QA names; the allow list wins over the deny list.
`AddConsumableIcons` puts every shipped consumable into the item icon table, where the UI looks its icon
up: the client's icon, or for an entry the client does not ship, the `Icon` its override states.

**The unsupported list.** `tools/database/unsupported_consumables.txt` names every catalogue consumable
that routes to no shape, as `<item id> <spell id> <name>: <reason>`, after the count routed to each
shape. It is a local baseline, not tracked: `TestEveryConsumableIsRoutedOrListed` writes it on the first
run and then names every line that moves. Once the moves are the intended ones, rewrite it:

```sh
UPDATE_UNSUPPORTED_CONSUMABLES=1 go test ./tools/database/ -run TestEveryConsumableIsRoutedOrListed -count=1
```

## Regenerating

`make db` extracts the client, then regenerates the spell store and then db.json. With
`tools/database/wowsims.db` already in place, the last two steps are:

```sh
go run ./tools/database/gen_spelldata
go run tools/database/gen_db/*.go -outDir=./assets -gen=db
```

The order matters: gen_db routes every consumable through the store compiled into it, so a change to
`ConsumableAllowList` or to a spell override needs the store rewritten first. The same gen_db run
rewrites the registrations below.

## How the effect is registered

`sim/core/consumes.go` keeps one effect per consumable id, and applies the registered effect of every id
the `ConsumesSpec` names. There are two ways in:

- `core.NewConsumableEffect(id, apply)` registers a hand-written effect.
- `core.NewDataDrivenConsumableEffect(id, apply)` registers an effect a builder in `sim/common/shared`
  made from the consumable's rows.

A hand-written effect wins over a data-driven one for the same id, whichever registers first, so the
two can live in different packages with no init order between them. Two hand-written effects, or two
data-driven ones, for one id panic. An id db.json does not carry registers nothing:
`core.MissingConsumableEffects()` lists it.

gen_db writes the data-driven registrations into `sim/common/forever`, one file per family, and
`common.RegisterAllEffects` runs them. The explosive and potion files keep a consumable their builder
refuses, commented out with the reason: Potion of Demonslaying 9224 ships through the allow list and
is applied by hand.

| family                                                         | file                                 | builder                                      |
| -------------------------------------------------------------- | ------------------------------------ | -------------------------------------------- |
| explosives that route as `damage`                              | `consumables_explosives_auto_gen.go` | `NewSpellDataExplosive`                      |
| potions (category 4) and conjured items (category 1153)        | `consumables_potions_auto_gen.go`    | `NewSpellDataPotion`, `NewSpellDataConjured` |
| other consumables that route as `heal`, `energize` or `absorb` | `consumables_potions_auto_gen.go`    | `NewSpellDataOnUseConsumable`                |
| `temporary enchant`                                            | `consumables_imbues_auto_gen.go`     | `NewTemporaryEnchantConsumable`              |
| `stat aura`, outside the potion and conjured categories        | `consumable_stats_auto_gen.go`       | `NewSpellDataStatConsumable`                 |

The potion builder runs one spell on the user: a heal, an absorb, a gain of mana, rage or energy, damage
to the user and the auras the parser attaches. A potion shares category 4's timer and is a major
cooldown only as the spec's combat potion; a conjured item is one only as the spec's conjured item; any
other is one whenever the spec names it. The stat builder puts an aura that outlasts the item's
cooldown up from the pull, and uses the item on its cooldown otherwise.

`applyConsumeEffects` handles three things itself: Potion of Demonslaying 9224, Gift of Arthas 9088, and
the pet's food, whose db.json stats go on the pet.

## Writing an effect by hand: the Natural flasks

Flasks of Natural Accuracy 274273, Aggression 274274, Precision 274275 and Swiftness 274276 cast
1293740-1293743. Each spell's first effect is 60 Stamina; its second is an `A_DUMMY` whose amount is the
bonus the text grants in Mount Hyjal, Hyjal Summit and the Barrow Deeps, and the auras after it state 0
and name what the bonus is (hit, crit, a lower chance to be dodged or parried, haste). The parser does
not attach an `A_DUMMY`, so the rows route as unsupported and no builder registers them.

Two steps make them work. `ConsumableAllowList` names the four, so db.json ships them although they
route as unsupported. `sim/common/forever/consumables_natural_flasks.go` registers each by hand, and
still reads every number from its row:

```go
func init() {
	for _, itemID := range []int32{274273, 274274, 274275, 274276} {
		core.NewConsumableEffect(itemID, func(agent core.Agent) {
			character := agent.GetCharacter()
			consumable := core.GetConsumableByID(itemID)
			row := spelldata.MustFind(consumable.ItemEffects[0].BuffId)

			config := spelldata.AuraConfig(row, spelldata.Label(consumable.Name))
			config.ActionID = core.ActionID{ItemID: itemID}
			config.BuildPhase = core.CharacterBuildPhaseConsumes
			spelldata.Permanent()(&config)
			aura := character.GetOrRegisterAura(config)

			spelldata.ParseEffects(character, aura, row, spelldata.Effects(1))
			attachNaturalFlaskZoneBonus(aura, row, row.EffectN(2).Average(character.Level))
		})
	}
}
```

The Stamina is the parser's, from the first effect. `attachNaturalFlaskZoneBonus` applies the dummy's
amount as the later auras name it, as though every fight is in those zones; `A_MOD_EXPERTISE` takes it
as a percent at four points of expertise each, since the text reads "5% reduced chance to be dodged or
parried". A hand-written effect for an id a builder does register takes that registration's place;
deleting the hand-written one hands the id back to the rows.

## Adding a consumable the rows do not support

The consumable needs a db.json entry and a hand-written effect.

**The client has it.** Add the id to `ConsumableAllowList`, as with the Natural flasks above. Potion of
Demonslaying 9224 routes to no shape (its 11406 applies `A_MOD_FLAT_SPELL_DAMAGE_VERSUS`) and ships
because the list names it:

```go
var ConsumableAllowList = []int32{
	9224, // Potion of Demonslaying
	// ...
}
```

An item outside the catalogue joins it through the same list. A catalogue consumable that
`ConsumableOverrides` names ships as well, with the override's fields merged onto the client's.

**The client does not have it.** Add the whole entry to `ConsumableOverrides`, since nothing else fills
it. Name, type, icon and every number the sim and the UI read come from there:

```go
var ConsumableOverrides = []*proto.Consumable{
	{Id: <item id>, Name: "<name>", Icon: "<icon file name>", Type: proto.ConsumableType_ConsumableTypeOther,
		CooldownDuration: <seconds>},
}
```

Then register its effect with `core.NewConsumableEffect` as above, and regenerate.

## Removing a consumable

Add the id to `ConsumableDenyList` and regenerate. The consumable leaves db.json, and a registration for
it then registers nothing. An id in `ConsumableAllowList` stays, so take it out of that list too.

```go
var ConsumableDenyList = []int32{
	10646, // Goblin Sapper Charge
}
```

## Correcting a number

A consumable's numbers are its spells' rows, so a number the client does not state goes in
`tools/database/overrides/spell_overrides.go` like any other spell's; the fields and the rule that
retires each are in [Spell Data: Overrides](spell_data.md#overrides). A row names the spell the
consumable casts, not the item:

```go
{30164, DurationMs, <ms>, "<the reason>", "<the source>"},
```

Permanent Lung Juice Cocktail 23715's 30164 is listed unsupported as `30164 states no duration for its
aura`; a `DurationMs` override on it lifts that refusal once the store and db.json are regenerated. An
override is refused when the client states the field itself. A number the client states but the sim
should not take is a hand-written effect for the consumable.

## Tests

- `TestEveryConsumableIsRoutedOrListed` (`tools/database`) keeps the unsupported list above, and
  `TestConsumablesRouteByTheirSpell` pins consumables to their shapes.
- `TestEveryShippedConsumableIsRegistered` (`sim/common`, `with_db`) fails with the ids of every db.json
  consumable that has neither a hand-written nor a data-driven registration, other than the two
  `applyConsumeEffects` handles itself. A consumable that routes to a shape no generator writes shows up
  here.
- Each builder has unit tests beside it in `sim/common/shared` (`consumable_*_test.go`), and the Natural
  flasks have theirs in `sim/common/forever`.
- The warrior suites select consumables from db.json; a regeneration that moves them shows up in their
  goldens.

## The browser path

`lib.wasm` is built without the database. Each request carries the consumables it needs in the player's
`database`, which `extendPlayerProtoWithMissingEffects` (`ui/sim/proto/proto_migration.ts`) fills from
the UI's db.json: every id field and the potion and conjured lists of the `ConsumesSpec`, and the item
behind every flag it sets. The flags are `core.ConsumableFlags` in `sim/core/consumes.go` (Goblin Sapper
Charge, the scrolls, Bogling Root and the pet's scrolls), which `gen_db -gen=go-to-ts` writes to
`ui/sim/proto/consumable_flags_auto_gen.ts`, so the sim and the UI read one table. A consumable travels
whole, its item effects and temporary enchant with it.

`TestRequestCarriesEveryConsumableThePresetSelects` (`sim/common`) runs a warrior once with every
consumable loaded and once with only those a request carries, and checks both apply the same;
`proto_migration.test.ts` checks the UI collects each flag's item. A new flag goes into
`core.ConsumableFlags`, then `make go-to-ts`.
