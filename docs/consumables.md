# Consumables

A consumable reaches the sim in two halves: its entry in `assets/database/db.json`, which the UI offers
and the sim looks the id up in, and an effect registered for its id, which the sim applies when the
character's `ConsumesSpec` names it. The generator writes the first from the client. The second is
built from the consumable's rows where a builder exists, and hand-written where the rows are not
enough or not right.

- [How a consumable gets into db.json](#how-a-consumable-gets-into-dbjson)
- [Regenerating](#regenerating)
- [How the effect is registered](#how-the-effect-is-registered)
- [Replacing a data-driven effect by hand](#replacing-a-data-driven-effect-by-hand)
- [Adding a consumable the rows do not support](#adding-a-consumable-the-rows-do-not-support)
- [Removing a consumable](#removing-a-consumable)
- [Correcting a number](#correcting-a-number)

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
first on-use row casts. Damage and heals are routed the way an item's on-use is; otherwise the first
effect of the spell, or of a spell it triggers on the user, that fits a shape decides it:

| kind                | what the spell does                                                    |
| ------------------- | ---------------------------------------------------------------------- |
| `damage`            | deals damage (Goblin Sapper Charge 10646: 13241)                       |
| `heal`              | heals the user (Major Healing Potion 13446: 17534)                     |
| `energize`          | restores a resource (Major Mana Potion 13444: 17531)                   |
| `stat aura`         | applies an aura the parser attaches (Flask of the Titans 13510: 17626) |
| `percent stat`      | multiplies a stat                                                      |
| `temporary enchant` | enchants a weapon (Brilliant Wizard Oil 20749: enchant 2628)           |
| `absorb`            | absorbs damage (Greater Fire Protection Potion 13457: 17543)           |

A consumable that fits no shape, or fits one the rows do not fully state (an aura with no duration, an
enchant slot the sim does not read), is unsupported, with a reason.

**What ships.** `ShipsConsumable` puts a catalogue consumable into db.json when it routes to a shape, or
when `ConsumableAllowList` or `ConsumableOverrides` names it. `ConsumableOverrides` is merged next, field
by field onto the client's entry, or as a new entry where the client has none. `KeepConsumable` then
drops what `ConsumableDenyList` names and the test and QA names; the allow list wins over the deny list.

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
`ConsumableAllowList` or to a spell override needs the store rewritten first.

## How the effect is registered

`sim/core/consumes.go` keeps one effect per consumable id, and applies the registered effect of every id
the `ConsumesSpec` names. There are two ways in:

- `core.NewConsumableEffect(id, apply)` registers a hand-written effect.
- `core.NewDataDrivenConsumableEffect(id, apply)` registers an effect a builder in `sim/common` made
  from the consumable's rows.

A hand-written effect wins over a data-driven one for the same id, whichever registers first, so the
two can live in different packages with no init order between them. Two hand-written effects, or two
data-driven ones, for one id panic. An id db.json does not carry registers nothing:
`core.MissingConsumableEffects()` lists it.

## Replacing a data-driven effect by hand

Register the hand-written effect for the same id, in an `init()` in `sim/common/forever`:

```go
func init() {
	// Major Healing Potion
	core.NewConsumableEffect(13446, func(agent core.Agent) {
		character := agent.GetCharacter()
		// register the spell and its major cooldown on character
	})
}
```

Nothing else changes: the data-driven registration for 13446 is dropped. Deleting the hand-written one
hands the id back to the rows.

## Adding a consumable the rows do not support

The consumable needs a db.json entry and a hand-written effect.

**The client has it.** Add the id to `ConsumableAllowList`. Potion of Demonslaying 9224 routes to no
shape (its 11406 applies `A_MOD_FLAT_SPELL_DAMAGE_VERSUS`) and ships because the list names it:

```go
var ConsumableAllowList = []int32{
	9224, // Potion of Demonslaying
	// ...
}
```

An item outside the catalogue joins it through the same list. A catalogue consumable that
`ConsumableOverrides` names ships as well, with the override's fields merged onto the client's.

**The client does not have it.** Add the whole entry to `ConsumableOverrides`, since nothing else fills
it. Name, type and every number the sim and the UI read come from there:

```go
var ConsumableOverrides = []*proto.Consumable{
	{Id: 23334, Name: "Cracked Power Core", Type: proto.ConsumableType_ConsumableTypeOther,
		CooldownDuration: int32(time.Hour.Seconds())},
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
