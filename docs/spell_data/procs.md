# Procs

The game files' proc chance column is not always a chance, so the generator bakes what the tooltip says
about it into the row as a `ProcChanceSource`. The four shapes:

|                     |                                                                                                                                                                                                                                  |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ProcChanceColumn`  | the tooltip renders `$h%`, so `SpellAuraOptions.ProcChance` is the roll                                                                                                                                                          |
| `ProcChanceEffectN` | the tooltip renders `$mN%`: the value of the effect at position `ProcChanceEffect` is the roll                                                                                                                                   |
| `ProcChanceAlways`  | the column reads 100 or 101 and the trigger clause states no chance at all, so the aura fires whenever its own condition is met                                                                                                  |
| `ProcChancePPM`     | the game files state no chance anywhere, or a 100/101 column sits beside a trigger clause saying the effect only sometimes happens ("Chance to strike your ranged target"), so the rate has to come from an override into `RPPM` |

Reading `ProcChance` directly is the bug `ProcChanceSource` exists to prevent: 100 and 101 are the
game files' "fires on its own condition" sentinel at least as often as they are a certainty.

```go
trigger := spelldata.ProcTrigger(&warrior.Character, spellData.Enrage.Rank(rank),
	func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		warrior.EnrageAura.Activate(sim)
	})
trigger.Name = "Enrage - Trigger"
warrior.MakeProcTriggerAura(trigger)
```

`spelldata.ProcTrigger(character, row, handler, opts...)` fills the listener the row describes: the
name and action id, the `Callback`, `ProcMask`, `Outcome` and `RequireDamageDealt` the decoder reads
out of `ProcTypeMask`, the internal cooldown, `CanProcFromProcs` and `ClassSpellsOnly` from the
attributes, the class mask the proc effect names, and the rate the source above points at. A row that
is a weapon proc aura also excludes the hits of spells flagged Suppress Weapon Procs.

The class flags of the spell that procs go on the trigger as `ProcSpellClassFlags`, and those are
what a `SPELLMOD_CHANCE_OF_SUCCESS` names: the rows its class mask reaches are the proc auras and the poison
spells themselves, not the spells that feed them. Improved Poisons 14113 names Instant Poison 8680, so
the temporary enchant's trigger built from 8680 rolls 20% plus 10 points. A trigger
`spelldata.ProcTrigger` builds draws once per hit: one with a manager rolls the manager's chance for
the hit's proc mask, and a mask the manager does not hear fails without a draw; one without rolls the
chance it states, and a stated 1 draws nothing. The unit's active `SpellMod_ProcChance_Flat` and
`SpellMod_ProcChance_Pct` mods go on whichever chance that is, the flat points first and the
percentage on the sum. The mods reach no spell, only a trigger that states `ProcSpellClassFlags`, which
`spelldata.ProcTrigger` fills from the row. A trigger built by hand reaches no talent unless it sets
them itself, as Judgement of Light does from its row. A class porting its poisons or another talented
proc builds it with `spelldata.ProcTrigger` and does not also add the talent's chance by hand, which
would count it twice.

The mods counted are the caster's wherever the trigger sits. `spelldata.ProcTrigger` sets `Caster`
to the character it is built for, and a trigger on the target names its caster: the paladin's
Judgement of Light passes the paladin, and Improved Judgement of Light 23564 raises the heal's
50%. A pet's trigger counts its owner's mods with its own, so a hunter talent parsed onto the hunter
reaches the pet. Frenzy 19621 names the proc of Tamed Pet Passive 20784, a row the store does not
hold, so the pet's trigger sets `ProcSpellClassFlags` from the talent's effect mask. Every proc chance
row attaches.

|                 |                                                                                                                                                               |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PPM(n)`        | a rate the game files do not carry. It clears the chance and binds a manager to the trigger's proc mask, so an option that narrows the mask has to come first |
| `Chance(f)`     | a rate the caller states outright, manager included                                                                                                           |
| `ChanceFrom(e)` | the chance an effect states, for a `$mN` the generator could not resolve                                                                                      |

A trigger that ends with neither a chance nor a manager panics naming the spell. A listener that fires
on every qualifying hit is never what a row with no stated rate means, and a procs-per-minute rate with
no proc mask to measure it on panics for the same reason - on a weapon proc, which hits count is the
weapon's to say.

## The decoder and the tooltip hints

`core.DecodeProcTypeMask(row.ProcFlags, row.ProcHint)` turns the game files' two proc words into a
`Callback`, a `ProcMask`, an `Outcome` and `RequireDamageDealt`, and names the bits it does not model
in `Unsupported`. `RequireDamageDealt` defaults to **true**, and a listener that fires on a dodge, a
parry, a miss or a block needs it false. Where the tooltip named that outcome the row carries
`ProcHintOutcomeTaken` and the decoder clears it already; a row whose wording yielded no hint - the
Battlegear of Wrath consume 23547 is one - has to be cleared by hand. Which outcome it is stays the
caller's either way, since no `ProcTypeMask` has a bit for any of them.

The mask states which hits reach the listener; it cannot state the condition around them, so the
generator reads that off the tooltip into `core.ProcHint`:

|                        |                                                                                                                                                         |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ProcHintCastTrigger`  | the tooltip names the cast itself, which turns a mask of the spell bits alone into `OnCastComplete` with no outcome                                     |
| `ProcHintCrit`         | the tooltip names a critical strike, which sets `Outcome = OutcomeCrit` and keeps the listener on the hit rather than the cast                          |
| `ProcHintHeals`        | the trigger clause names healing or an unrestricted spell, which is the evidence that a helpful-spell bit carries a real trigger rather than a leftover |
| `ProcHintPureHeal`     | the trigger is healing only, so the damage callbacks come off                                                                                           |
| `ProcHintNamedAbility` | the clause names one ability ("your Shock spells"), which no proc mask can state                                                                        |
| `ProcHintOutcomeTaken` | the clause names an outcome the mask has no bit for, which clears `RequireDamageDealt`                                                                  |

The last two are shapes the decode reads past rather than models. A trigger is still built for them
and for the bits the decoder names as unsupported: a listener that hears fewer hits than the game files
state is deliberately the narrower one.

A class mask naming **another class's** family is dropped rather than obeyed. An item every class can
wear states the filter of the one class it was written for, and on anyone else that mask matches
nothing and would silence the listener instead of narrowing it. A family the game files state for no class
and the empty mask are evidence of nothing and are left alone.

## The shapes, as the warrior reads them

- **Enrage** (`sim/warrior/talents_fury.go`) is the column: its tooltip renders `$h%`, so the 30 in
  the column is a real roll on damage taken. Only the trigger's name is the caller's, because the
  driver and the buff share one.
- **Shield Specialization** (`sim/warrior/talents_protection.go`) is the effect ladder: the column
  reads 100 and the roll is effect 2. Its tooltip names an outcome, which the row carries as
  `ProcHintOutcomeTaken`, so the decoder already clears `RequireDamageDealt` - a block deals no damage.
  Which outcome it is stays the caller's: `Outcome = OutcomeBlock`.
- **Flurry** (`sim/warrior/talents_fury.go`) is the no-roll shape: the row states "always" and the
  crit is the condition. The buff row supplies the duration and the three charges through
  `AuraConfig`, the haste comes off the talent's ladder, and the mask stays the caller's because the
  row's reaches ranged and spell hits while the tooltip says melee. Where the row and the tooltip
  disagree on a number - 12966's flat 30 against the ladder's 25 at five points - the decision is
  written at the call site.
- **Lightning Shield** 324 is the same no-roll shape on a row nothing ports yet: 100 in the column,
  three charges, a 3.5 second internal cooldown, and its damage spell reachable only as a tooltip
  reference while its effect triggers the dispatcher every rank shares.

## Item and enchant procs

An item, enchant or set proc is registered from two spell ids and a shape:

```go
shared.NewSpellDataDamageProc(shared.SpellDataProc{TriggerSpellID: 7711, BuffSpellID: 7712},
	[]shared.ItemVariant{
		{ItemID: 17111, ItemName: "Blazefury Medallion"},
	})
```

`NewSpellDataProc` is the same for a proc whose buff is an aura. What the listener
hears, how often it fires and its internal cooldown come from the trigger's row; how long the buff
lasts and how it stacks come from the buff's; only the stats come from the item's effect entry, where
the sim's item level scaling lives. `IsWeaponProc` states the one shape no row describes: the game
casts a chance-on-hit effect off the weapon's hit without consulting a proc mask, and there the 100/101
sentinel is never a rate.

A buff the game files state as a percentage of a stat (`A_MOD_PERCENT_STAT`, misc the stat index, -1 all
five) has no effect entry, which carries flat stats only, so the generator routes it as auras:
`NewSpellDataAuraProc` parses the buff's row, whose multiplier works through a dynamic stat dependency
and reports what it adds and removes to the temporary stats listeners. Insight 8216's 1299796 doubles
Spirit for 10 s. The aura proc registers the wearer's buff with the item swap, so an enchant's buff
drops with its weapon or shield, and, where the buff moves stats, with the stat-proc APL values. The
game files have no item proc of this shape. `ParseStatic` reads the same aura as a multiplier applied once,
which reports nothing.

An enchant's procs are read one per slot of its `SpellItemEnchantment` row that casts a combat spell
(Effect 1) or hangs an equip aura off a hit (Effect 3), and at most one of them registers. A combat spell's chance, where the game files state
one, is on the enchantment rather than on the spell - `EffectPointsMin`, Fiery Blaze's 15 - and the
store writes it onto the spell's row as its `ProcChance` column, noted beside the row. Every
enchantment casting one spell states the same chance for it; the generator stops on one that does
not. The chance is rolled on the hits of the enchanted weapon only
(`shared.statedWeaponProcChance`). An equip aura's own description is empty, so the store reads
the description of the spell granting the enchant in its place, for what no proc mask can state: the
named ability, outcome-taken, attack-dodged and attack-parried hints, and whether its 100 is a
sentinel ("often", "sometimes", "a chance to"), which makes it `ProcChancePPM`. The grant's cast,
crit and heal wording is left alone: which hits feed the aura is its own mask's to say.
A combat spell and an aura applying the same spell (Crusader) register once, as the combat spell.

`spelldata.ItemProcUnsupportedWithRate(trigger, isWeaponProc, itemPPM)` is the single decision about
whether the rows say enough; `itemPPM` is the rate a weapon's item effect states, 0 where it states
none. The generator calls it when it writes the item files, the audit calls it, and a test pins it, so a proc the generator emits is one the sim can build and a proc it comments out carries the
reason in the comment. `tools/database/unsupported_procs.txt` is the audit's census of every proc an
item, enchant or set can reach and what the rows refuse; `sim/common/forever/registered_effects.txt`
lists the item and enchant effects that do register. Neither is tracked: each is a local baseline its
test writes on the first run and then compares against, naming every line that moved.
