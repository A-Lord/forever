# Reading wowsims.db

Routed from `## Tables the generator does not ship` in `SKILL.md`: what the tables in `tools/database/wowsims.db` hold
and how a script reads them.

- **Empty tables are real.** The game files ship no gem system, zero `SpellScaling` columns, no `SpellLearnSpell`, no
  procs per minute (`SpellProcsPerMinuteID` 0), no Jewelcrafting or Inscription, and `ItemSparse.AllowableClass` is -1
  on class-only items such as Atiesh.
- **Array columns hide generated columns.** `db2tool` adds a `<Column>_<n>` column per element
  (`EffectSpellClassMask_0..3`, `ImplicitTarget_0`, `Attributes_N`; `tools/db2tool/sqlite/schema.go`) that
  `pragma table_info` hides and `select *` returns. Take column names from `cursor.description` or select an explicit
  list.
- **`SkillLineAbility.RaceMasks` is `[low, high]`** words indexed by `ChrRaces.PlayableRaceBit`, so Skyborne's bits
  32 and 33 land in the high word as 1 and 2.
