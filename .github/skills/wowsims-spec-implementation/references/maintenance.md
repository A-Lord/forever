# Keeping wowsims-spec-implementation true

| You changed…                                                     | Update                           |
| ---------------------------------------------------------------- | -------------------------------- |
| how `tools/database/gen_spell_data.go` skips Engrave-only spells | "Deciding what is actually live" |
| the `spelldata-check` target in `makefile`, or the test tags     | "Verifying"                      |
| a `##` heading this skill names in another skill                 | the pointer that names it        |

Then run `sh .github/skills/check_paths.sh .github/skills/wowsims-spec-implementation`.
