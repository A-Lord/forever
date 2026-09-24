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
- [What the settings offer](#what-the-settings-offer)
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
data-driven ones, for one id panic. An id db.json does not carry registers nothing.

`GenerateConsumableRegistrations` (`tools/database/gen_consumable_registrations.go`) writes the
data-driven registrations into `sim/common/forever`, one file per family, and
`common.RegisterAllEffects` runs them. It picks each consumable's builder once, by its type, the
category its on-use shares and how its spell routes, in the order of the table. The explosive and
potion files keep a consumable their builder refuses, commented out with the reason: Potion of
Demonslaying 9224 ships through the allow list and is registered by hand. Each builder states what it
refuses beside itself (`shared.ConsumableUnsupported`, `DamageConsumableUnsupported`,
`StatConsumableUnsupported`), and the router reads the same checks.

| family                                                         | file                                 | builder                         |
| -------------------------------------------------------------- | ------------------------------------ | ------------------------------- |
| `ConsumableTypeImbue`                                          | `consumables_imbues_auto_gen.go`     | `NewTemporaryEnchantConsumable` |
| `ConsumableTypeExplosive` that routes as `damage`              | `consumables_explosives_auto_gen.go` | `NewSpellDataExplosive`         |
| potions (category 4) and `ConsumableTypeConjured`              | `consumables_potions_auto_gen.go`    | `NewSpellDataOnUseConsumable`   |
| other consumables that route as `heal`, `energize` or `absorb` | `consumables_potions_auto_gen.go`    | `NewSpellDataOnUseConsumable`   |
| other consumables that route as `stat aura`                    | `consumable_stats_auto_gen.go`       | `NewSpellDataStatConsumable`    |

The potion builder runs one spell on the user: a heal, an absorb, a gain of mana, rage or energy, damage
to the user and the auras the parser attaches. A potion shares category 4's timer and is a major
cooldown only as the spec's combat potion; a conjured item is one only as the spec's conjured item; any
other is one whenever the spec names it. The builder tells them apart by the consumable's type. Damage a
consumable deals its own user, a sapper's or Demonic Rune's, always lands and can crit. The stat builder
puts an aura that outlasts the item's cooldown up from the pull, and uses the item on its cooldown
otherwise.

`shared.NewVersusPotion`, which `sim/common/forever/consumables_versus.go` registers Potion of
Demonslaying 9224 with by hand, is the potion builder with its spell 11406's `A_MOD_*_VERSUS` auras put
on the potion's aura: 185 attack power and 108 spell damage against the creature types their misc
value masks (4, Demon), for the row's 120 s, on category 4's 120 s timer. The same file registers
Consecrated Sharpening Stone 23122 with `shared.NewVersusTemporaryEnchant`: core puts its enchant 2684
on each hand whose weapon the masks take, and the enchant's equip spell 28893 grants 99 melee and 99
ranged attack power against Undead once per such hand, since each weapon carries its own equip aura
and 28893 marks no aura unique. Blessed Wizard Oil 23123 goes the same way: its enchant 2685's equip
spell 28890 grants 58 spell damage against Undead (`A_MOD_FLAT_SPELL_DAMAGE_VERSUS`, mask 32). `consumables_gift_of_arthas.go` registers Gift of Arthas 9088.

`applyConsumeEffects` handles one thing itself: the pet's `pet_extra_ids`, whose db.json stats go on
the pet.

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
var NaturalFlasks = []int32{274273, 274274, 274275, 274276}

func init() {
	for _, itemID := range NaturalFlasks {
		core.NewConsumableEffect(itemID, func(agent core.Agent) {
			character := agent.GetCharacter()
			consumable := core.GetConsumableByID(itemID)
			row := spelldata.MustFind(consumable.OnUse().BuffId)

			config := spelldata.AuraConfig(row, spelldata.Label(consumable.Name))
			config.ActionID = core.ActionID{ItemID: itemID}
			config.BuildPhase = core.CharacterBuildPhaseConsumes
			spelldata.Permanent()(&config)
			aura := character.GetOrRegisterAura(config)

			spelldata.ParseEffects(character, aura, row, spelldata.Effects(1))
			bonus, err := NaturalFlaskZoneBonusOf(row, row.EffectN(2).Average(character.Level))
			if err != nil {
				panic(err)
			}
			aura.AttachStatsBuff(bonus.Stats)
			if bonus.AttackHastePercent != 0 {
				aura.AttachMultiplyAttackSpeed(1 + bonus.AttackHastePercent/100)
			}
			if bonus.CastHastePercent != 0 {
				aura.AttachMultiplyCastSpeed(1 + bonus.CastHastePercent/100)
			}
		})
	}
}
```

The Stamina is the parser's, from the first effect. `NaturalFlaskZoneBonusOf` reads the dummy's amount
as the later auras name it, as though every fight is in those zones; `A_MOD_EXPERTISE` takes it as
`stats.ExpertisePercent`, since the text reads "5% reduced chance to be dodged or parried". A
hand-written effect for an id a builder does register takes that registration's place;
deleting the hand-written one hands the id back to the rows.

The settings offer a consumable by the stats db.json lists for it, and the parser lists only the
Stamina. `AddNaturalFlaskZoneBonusStats` (`tools/database/consumable_zone_bonus.go`) adds the same
`NaturalFlaskZoneBonusOf` to each flask's db.json stats at the rating one percent takes (Accuracy 50 melee and spell hit rating,
Aggression 56 melee and spell crit rating, Precision 50 expertise rating at
`core.ExpertiseRatingPerExpertisePercent` (10), Swiftness 50 melee and spell haste rating). The sim reads the rows, not those stats.

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
		ItemEffects: []*proto.ItemEffect{{BuffId: <spell id>, Effect: &proto.ItemEffect_OnUse{
			OnUse: &proto.OnUseEffect{CooldownMs: <ms>}}}}},
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
  consumable that has neither a hand-written nor a data-driven registration. A consumable that routes to a shape no generator writes shows up
  here.
- Each builder has unit tests beside it in `sim/common/shared` (`consumable_*_test.go`), and the Natural
  flasks have theirs in `sim/common/forever`.
- The warrior suites select consumables from db.json; a regeneration that moves them shows up in their
  goldens.
- `ui/sim/settings/consumables.test.ts` pins the pickers and the relevance rules below,
  `ConsumesPicker/utils.test.ts` the settings rows, and `consumes_migration.test.ts` the migration of
  settings saved before api version 2.

## What the settings offer

The settings render every consumable db.json ships from its own fields; nothing in the UI names a
consumable. `ui/sim/settings/consumables.ts` puts each in a picker by its type: `ConsumableTypeConjured`
the conjured item, `ConsumableTypePotion` the combat potion, then the flask, the two elixir pickers
(`ConsumableTypeElixir`, Battle or Guardian), `ConsumableTypeFood` the food, the explosive and the
imbues. Everything else is picked any number at a time into the spec's `extra_ids`: the scrolls
in their row, the rest in the miscellaneous row, and an explosive with a cooldown of its own beside the
shared one (Goblin Sapper Charge 10646) beside the explosive picker. The pet's scrolls go into
`pet_extra_ids`.

The generator writes the type (`Consumable.ToProto` in `tools/database/dbc/consumable.go`) from the
category the first on-use row shares, which is also the consumable's `category_id`: the item row's, or
where it states none, its spell's. An imbue or an explosive is typed by what it does; category 4 makes a
potion (Potion of Demonslaying 9224 is an Elixir-subclass item and only its spell 11406 names the
category) and 1153 a conjured item (Demonic Rune 12662 is not even a Consumable-class item); food and
drink categories 11 and 59 make food of what its subclass leaves neither a potion, a flask nor an elixir;
everything else takes its item subclass.

`isConsumableRelevant` then decides whether the player sees it:

- the class, profession and race the client allows it (`class_allowlist`, `required_profession`,
  `race_allowlist`, the last read from `ItemSparse.AllowableRace` by the generator);
- a stat the spec weighs (its `epStats` and `consumableStats`; Stamina brings health and the
  resistances, Attack Power flat physical damage), a resource it spends, or, for a consumable granting
  neither, an effect of another kind (damage, an absorb, a weapon enchant);
- for an imbue, the hand's item against the enchant's weapon subclass and inventory type masks, so a
  shield coating shows on a shield.

Of the ranks of one name (Scroll of Agility .. IV) the settings offer the highest numeral among those
the player is offered, so a top rank the class, the hand or the profession rules out leaves the rank
below it in the picker; the pet's scrolls, offered by spec, keep the highest numeral. The UI keeps
the rows and their text, the flask and elixir exclusion, the Windfury Totem rule for the main-hand imbue
and which specs have a pet. `Player.getConsumes(true)` sends the relevant potions and conjured items as
the spec's lists.

`gen_db -gen=db` writes `ui/features/settings/model/consumables_auto_gen.ts` in the same pass as db.json:
one named export per shipped consumable, carrying its item id (`MightyRagePotion = { id: 13442 }`).
Presets name consumables through it, so a preset cannot name one db.json does not ship. A name several
items share takes each item's id (`MajorHealthstone9421`).

## The browser path

`lib.wasm` is built without the database. Each request carries the consumables it needs in the player's
`database`, which `extendPlayerProtoWithMissingEffects` (`ui/sim/proto/proto_migration.ts`) fills from
the UI's db.json: every id field and every id list of the `ConsumesSpec` (the potions, the conjured
items, `extra_ids` and `pet_extra_ids`). A consumable travels whole, its item effects and temporary
enchant with it.

`TestRequestCarriesEveryConsumableThePresetSelects` (`sim/common`) runs a warrior once with every
consumable loaded and once with only those a request carries, and checks both apply the same.

Settings saved before api version 2 name the Goblin Sapper Charge, the scrolls, Bogling Root and the
pet's scrolls by flags, drums by an enum and the pet's food by an id. `REMOVED_CONSUMES_FIELDS`
(`ui/sim/proto/consumes_migration.ts`) names each removed field with the item a set flag selected and
the number it had in the two layouts a link may carry: version 1's, and the one from before the
version restarted at 1, whose links are stamped 16. The same file turns a link's (binary, by field
number, `sim_links.ts`) or a saved blob's (JSON, by name) flags into item ids and the spell ids the
explosive and imbue fields held into the items that cast them; `migrateLegacyConsumes` does it for
both JSON shapes, saved settings and a sim's settings, and drops the party's drums.
`TestRemovedConsumesFieldsStayReserved` (`sim/core`) keeps every one of those fields reserved in
`ConsumesSpec`, version 1's by number as well.
