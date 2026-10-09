---
name: wowsims-conventions
description: 'Use whenever writing or editing anything that lands in the WoWSims Forever repository or on its GitHub: Go or TypeScript comments, docs, skills, test messages, commit messages, PR titles and bodies, issue text. Covers how comments are written, the words the repo uses for the game data, and terms that stay out of committed text.'
---

# Writing conventions

These apply to every file and every piece of GitHub text, in Go and TypeScript alike. The subject skills
(`wowsims-spells`, `wowsims-spec-implementation`, `wowsims-ui`, ...) assume them.

## Comments

- **State what the code does now.** No history and no comparison with another version: not "this used
  to...", "TBC fired...", "was a TBC build", "now reads...". Where the old behaviour is needed to make sense of
  the code, keep the fact that justifies it ("232 of 235 talents list ids the game files do not ship") and drop
  the narrative. History lives in git.
- **Comment what the code cannot say.** A trap, where a magic number comes from, a game rule the code
  implements. Why a bug happened, and what changed to fix it, goes in the commit message, not above the function.
- **Short where the neighbours are short.** Struct fields with one-line trailing comments get a one-line trailing
  comment, not a paragraph block:

    ```go
    ClassSpellMask      int64      // The sim's own custom spell masks, not from the game files.
    ClassFlags          ClassFlags // The spells this trigger fires on (EffectSpellClassMask in the game files); with ClassSpellMask, both must match.
    ProcSpellClassFlags ClassFlags // The proccing spell's class flags, which proc chance talents match (e.g. Improved Poisons on Instant Poison).
    Caster              *Unit      // Whose proc chance mods apply, when not the aura's unit (e.g. Judgement of Light on the target).
    ```

- **Do not name what the comment sits on.** A comment never opens with or repeats the name of the function,
  type or field it documents: "The first row under an effect id that fits the item", not "FittingEnchantRow
  returns the first row…". This holds for Go doc comments too, against Go's usual convention.
- **State the rule, not the site it came from.** Do not name SimC, Wowhead or a JS mapping as the source of a
  name or value ("one SimC names", "commented out in the JS mapping"); write the rule ("one with no known name is
  `A_<id>`"). Existing Wowhead spell links above a spell's registration are fine; do not add attribution prose.

## Words

- **"The game files", never "the client".** The data extracted from WoW (the DB2 tables, `wowsims.db`, the
  spell store's source) is "the game files" or "game files data": "EffectSpellClassMask in the game files", "the
  game files state no chance", "a game files row". The game version is the "game build" (1.60.1.x). What the
  running game does is "in game". "Client" stays only where it means something else: an HTTP or API client, the
  browser, DOM properties (`clientX`), CASC paths (`dbfilesclient`), and existing identifiers.
- **Two kinds of spell mask.** `ClassSpellMask` is the sim's own custom masks, which a class defines
  itself (e.g. `sim/paladin/spell_masks.go`). `ClassFlags` are the class flags from the game files (family plus a 128-bit mask). Say which
  one a sentence means.
- **No "golden".** Committed text and GitHub text say ".results files" or "expected results" (the sim tests), or
  "the state-snapshot capture" (UI), never "golden" or "goldens". Item names such as Golden Scale Coif are
  unrelated.

## Before committing

- Grep the added lines for `client`, `golden` and `used to`, and check each hit against the rules above.
- `npx oxfmt . --check` covers markdown in `docs/` and `.github/`, so run it after editing either.
