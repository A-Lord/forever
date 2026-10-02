# Regenerating and checking

```
go run ./tools/database/gen_spelldata              # rewrite every generated spell data file
go run ./tools/database/gen_spelldata -classes=druid  # take the druid from the database, keep every other class as committed
go run ./tools/database/gen_spelldata -check       # name the ones that are stale, write nothing
go run ./tools/database/gen_spelldata -unchecked   # write without the type-check below
make spelldata-check                               # the same check
```

The generator reads `tools/database/wowsims.db` and writes all of it in one pass:
`sim/core/spelldata/spells_auto_gen.go`, `sim/core/dbcenums/forms_auto_gen.go`, a
`sim/<class>/spell_data_auto_gen.go` of ladders into the store for every class, and the raid buffs,
which [Buffs and debuffs](buffs-and-debuffs.md) describes. It is its own binary rather than a mode of `gen_db` on purpose:
`gen_db` imports the sim and the sim reads these files, so a stale one would stop the generator that
fixes it from compiling. For the same reason nothing is written until all of it type-checks: the
rendered bytes go to a staging directory first and are compiled through `go build -overlay` in the
place of the committed files, and a failure leaves the tree exactly as it was and prints what the
compiler said. `-unchecked` skips that build, for a class file whose call sites cannot type-check
against what it will state until they move - the file is written first and the call sites follow;
`-check` closes the loop once the package builds again. `make db` runs the store before `gen_db` for the same ordering reason - `gen_db`
classifies every item and enchant proc out of the store compiled into it.

Nothing lists which spells to generate. The class files walk `dbc.Classes`, take each class's own skill
lines and emit a family for every spell whose subtext reads `Rank N`; the store starts from what those
families name, from the item, enchant and set-bonus tables, and from the talent trees, and closes over
everything those reach. A Season of Discovery rune ability is left out: a family every spell of which
an `Engrave <slot> - <rune>` spell reaches, with no trainer or level row outside the rune lines and no
talent reaching it, as Shadow Cleave is. So is the undescribed dummy a rune's action bar override
replaces, whatever its rows, which is Metamorphosis : Demon Charge. A rune ability a trainer also
teaches, such as Victory Rush or Incinerate, stays. A family that could not be resolved, or a rune's,
is named in the `// Not generated:` comment at the head of the class file.

`assets/db_inputs/spell_store_inputs.json` holds the game files rows the store was built from, committed
beside it, and every `SpellShapeshiftForm` row whole alongside them, since a form names no spell for
the store's closure to reach it by. It is gzip-compressed JSON despite the name, like everything under
`assets/db_inputs/dbc` - `zcat` it to read it. It exists so the store, the buffs and the shapeshift
forms can all be rebuilt without the game files database, which is gitignored and comes from a local WoW
install, and `-check` compares it too: a capture that no longer matches the database would render the
committed files all the same, so nothing else would notice it going stale.

What guards the outputs:

- `TestStoreRegeneratesFromTheCommittedInputs` in `tools/database` re-derives the closure, the hand
  links, the curves, the tooltip hints, the overrides and the emitter from that capture and asserts the
  committed store is what comes out. No build tag and no database, so it runs everywhere.
  `TestBuffFilesRegenerateFromTheCommittedInputs` does the same for the buff files, and
  `TestFormsRegenerateFromTheCommittedInputs` for `sim/core/dbcenums/forms_auto_gen.go`.
- `sim/core/spelldata/snapshot_test.go` pins the shape and the counts of the committed store, and a
  handful of rows read out of the game files by hand, so a regeneration that moves a number says so there
  instead of in a sim result.
- The repository's `pre-commit` hook runs `-check` when the database is present and the commit touches
  the generator or one of its outputs.

The .results files are the last check and the one that costs time: run the suites of the class that changed, and
diff the `.results` against the `.results.tmp` with a local diff helper or by hand. A port that
was meant to be mechanical and moves a .results number is a wrong port, not a new baseline.

Taking only some classes or the racials from a new game build is described in
[Taking a new game build one spell data group at a time](partial-updates.md).
