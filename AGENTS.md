# Agent instructions

WoWSims Forever: a Go simulator (`sim/`) with a React frontend (`ui/`), built on data extracted from the game
files (`tools/database`, `sim/core/spelldata`).

**Rule: tests and stories never add code to the sim or the UI.** No export, prop, parameter, branch, flag or
hook goes into `sim/` or `ui/` only so a unit test or a Storybook story can reach or cover something. A test
or story works through the code's real surface, and a fixture or stub it needs lives with the tests.

The project's skills live in `.github/skills/`. Read the one that matches the task before starting, and always
read `wowsims-conventions` before writing comments, docs, test messages, commit messages or GitHub text.
`make skills-check` verifies every path and skill name the skills mention.

| Skill                         | Use it for                                                                                     |
| ----------------------------- | ---------------------------------------------------------------------------------------------- |
| `wowsims-conventions`         | How comments, docs, commits and PR text are written; "game files", not "client"; banned terms  |
| `wowsims-spells`              | The map: which spell data skill to read, where things live, when the sim and the row disagree  |
| `wowsims-spell-store`         | Reading the spell store: ladders, ranks, effects, hand reads and resolvers                     |
| `wowsims-abilities-auras`     | Building abilities, auras, dots, talents and buffs from their rows                             |
| `wowsims-procs`               | Proc triggers: chances, PPM, ICDs, charges, item, enchant and set procs                        |
| `wowsims-consumables-items`   | Items, enchants and consumables in the database, and the effects registered for them           |
| `wowsims-db-pipeline`         | Regenerating the store and the database, overrides, hotfixes, a database behind master         |
| `wowsims-core-mechanics`      | What `sim/core` does around a spell: cast order, spends, scaling, targets, hooks, auras, state |
| `wowsims-apl`                 | Rotations (APL): preset lines, timing, values, why a cast never fires                          |
| `wowsims-testing`             | Proving Go work: the CI gates, .results files, what a test asserts, runtime proofs             |
| `wowsims-spec-implementation` | Picking up a `[Spec]` tracking issue and porting a class                                       |
| `wowsims-review`              | Reviewing a spell change, your own before committing or a pull request, against the rows       |
| `wowsims-ui`                  | The frontend under `ui/`                                                                       |
| `wowsims-bulk-sim`            | Bulk sim, locally, on the server and in the browser                                            |
| `wowsims-reforge-optimizer`   | The gear optimizer (EP, stat caps, HiGHS)                                                      |

A ruling, trap or workflow fact learned while working goes into the subject skill, or its `references/rulings.md`, in
the same change.
