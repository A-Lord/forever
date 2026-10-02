# Overrides

`tools/database/overrides/spell_overrides.go` holds the numbers the store carries that the game files
do not state. One row per number:

```go
{16928, PPM, 1, "Annihilator: ProcChance 101 sentinel; TBC 1 PPM, unverified on Forever", "tbc-carryover"},
```

`Reason` and `Source` are not decoration: they are what the next reader weighs the number against, and
the generator refuses a row without a reason. `Source` names where it came from - `tbc-carryover`,
`tooltip`, `wcl:<report>/<fight>`, `issue #N`.

Every field names a rule the generator checks before it writes the value, so an override outlives its
reason no longer than the next regeneration:

|                                   |                                                           | stale when                               |
| --------------------------------- | --------------------------------------------------------- | ---------------------------------------- |
| `PPM`                             | procs per minute, onto `Spell.RPPM`                       | the row states a `SpellProcsPerMinuteID` |
| `FlatThreat`                      | threat the tooltip states without a number                | the spell gains a threat effect          |
| `APCoefDirect` / `APCoefPeriodic` | an attack power coefficient the game files keep in script | the game files state one on that effect  |
| `ProcChancePct`                   | a whole percentage onto `Spell.ProcChance`                | the tooltip states a chance of its own   |
| `DurationMs`                      | an aura duration                                          | the game files state a `SpellDuration`   |
| `Hint`                            | the `ProcHint` bits the tooltip's wording did not yield   | -                                        |

Two more refusals have nothing to do with the field: an override naming a spell the store does not
carry, and two overrides of the same field on one spell. Every one that is applied leaves an
`// override: <field> <value> -- <reason>` comment on the row it wrote to, so reading the generated
store says which numbers are not the game files'.

An item or enchant proc refused as `states no rate` loses that refusal once a `PPM` override sits on
the spell it is routed through - the `TriggerSpellID` of its commented-out registration, the id on its
`// trigger N` line - and registers where it was the only one:

- A combat enchant (Effect 1) or a chance-on-hit item effect: the combat spell itself - Fiery Weapon's
  13897, Unholy Weapon's 20006 - not the spell that grants the enchant. The rate is measured on the
  hits of the weapon carrying it (`NewDynamicLegacyProcForEnchant(id, ppm, 0)`, or `...ForWeapon` for an
  item).
- An equip aura (Effect 3): the aura carrying the proc trigger - Revelation's 1248806 - not the spell it
  triggers (1248808) nor the grant (1248805). The rate is measured on the aura's own proc mask, so an
  aura whose flags decode to no mask stays refused as `no proc mask to measure its rate on`.

A procs-per-minute enchant proc rolls on weapon hits only: its mask keeps its melee and ranged bits,
and a spell or heal hit never procs it. `dpmForMask` strips the rest for every enchant, and
`spelldata.EnchantAuraUnsupported` refuses a rate on an equip aura that hears only spells as
`a procs-per-minute rate hears no weapon hits in this mask`, so it stays listed rather than
registering to never fire. An enchant the game does let spells proc is an exception to state there,
by enchant; there is none today. A stated chance (Fiery Blaze 36's 15%, Insight 8216's 35%) is not a
rate and hears what its row says.

Revelation 8217 stays listed. Its rate is scripted and the game files do not state it: trigger 1248806's
`ProcChance` 100 is the sentinel beside the tooltip's "a chance", and its effect entry resolves no stats
from 1248808.

A rate follows the weapon the enchant sits on. A weapon enchant, ranged ones included
(`NewDynamicLegacyProcForEnchantWithMask`), rolls only on the hand carrying it, at that weapon's speed;
an item swap moves it with the weapon. An enchant on no weapon - armor, a cloak, a shield or a
held-in-off-hand item - is priced off the main hand (`NewLegacyPPMManager`, which prices every mask but
the off hand's and the ranged one's at the main hand's speed).

Where a combat spell and an aura apply the same spell (Crusader), the slot whose row states a rate is
the one kept, so the override goes on the combat spell to keep it the combat spell. After adding a row,
run `gen_spelldata` and then `gen_db`, which classifies the procs out of the store compiled into it.
`TestEnchantProcRoutingTakesAPPMOverride` in `tools/database` and `TestSpellDataProcTakesAPPMOverride`
in `sim/common/shared` pin both halves on those three rows with a rate that exists only in the test.
`TestEnchantAuraPPMFollowsItsWeapon` beside the latter pins the routing by slot, and
`TestEnchantPPMFollowsTheEnchantedWeapon` in `sim/core` the move across an item swap.

`overrides.AreaBonuses` is the second table in the same file, for an effect whose tooltip says it is
doubled in some kind of area while the game files state no companion row for it ("This effect is doubled
in Volcanic areas" on Molten Fury). A row names the AreaGroup ids any of which counts, the factor on
the effect's amounts and the factor on its duration, onto `Spell.AreaBonusGroups`, `AreaMultiplier`
and `AreaDurationMultiplier`; `row.AreaBonus(&character.Env.Encounter)` reads them against the
encounter's area set, and is stale once the row states a `RequiredAreasID` of its own. The on-use stat
actives `shared.NewSimpleStatActive` registers read it; a proc chance the tooltip says is doubled has
no reader yet.

A consumable's spells take these overrides like any other spell's; the lists that add, remove or
replace a consumable itself are in [Consumables](../consumables.md).

`tools/database/overrides/extra_spells.go` is the other hand-kept list: spells the generator
force-includes although no class, item, enchant or set reaches them. It is empty today. Each entry
states a reason and a source like an override, and the generator refuses one without a reason. The
extras are added while the store renders, so adding one without regenerating fails the check rather
than shipping a store that does not match its source.
