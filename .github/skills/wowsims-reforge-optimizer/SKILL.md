---
name: wowsims-reforge-optimizer
description: 'Use when working on the WoWSims gear optimizer (the "reforge optimizer" in code; there is no reforging), the /reforgeOptimizeAsync endpoint, HiGHS solver integration, EP/cap logic, or softcap breakpoints.'
argument-hint: 'Describe the optimizer bug, fixture, or behavior to work on.'
---

# WoWSims Gear Optimizer Guide

Comments, docs, commit messages and PR text follow `wowsims-conventions`: "game files", not "client", no history in comments, no "golden".

## Scope

- Core optimizer behavior in sim/core/reforge_optimizer ("reforge" is only the code's name; there is no reforging).
- Endpoint path /reforgeOptimizeAsync and worker integration.
- HiGHS-backed MIP solving under stat caps.

The optimizer picks one static enchant per listed slot to maximize EP within the stat caps.

- Candidates come from `ReforgeOptimizeRequest.enchant_options`, built by ui/features/reforge/model/enchant_options.ts. Per equipped, unfrozen slot it lists the enchants the player can use on that item (class and Enchanting gated, bulk's fit rule `core.EnchantAppliesToItem`) that are static (`proc_kind` None, with stats, pseudo-stats or weapon damage) and worth EP > 0 on the optimizer's pre-cap weights.
- A slot whose item can take a proc enchant the sim implements (a combat proc for every spec, a survival proc for tanks) is skipped and listed in the panel's Enchants section, with a link to the batch sim. A proc enchant in `MISSING_ENCHANT_EFFECTS` (ui/sim/constants/missing_effects_auto_gen.ts) blocks nothing; it has no stats, so it is never a candidate either. Every melee weapon takes Crusader, so the main and off hand are skipped when they hold one.
- "Use enchant presets" (`ReforgeSettings.use_enchant_presets`, off by default; the UI reads it, Go does not) swaps the candidate source for the checked enchant presets, the same presets and checks the batch sim uses (`enchantPresetPool` in ui/features/enchant-presets/model/enchant_presets.ts; Finger 1 lists both rings, Weapons both hands). Per equipped, unfrozen slot: an item that takes enchants but none the presets list skips the slot and lists it under "No checked enchant preset lists these slots:" (an item that takes none, such as a wand, is silent); a listed proc the sim implements, combat or survival, skips it for every spec; otherwise the candidates are exactly the listed static enchants, with no EP filter. While it is on and no preset is checked, Optimize Enchants is disabled.
- `proc_kind` comes from gen_db (tools/database/dbc/enchant.go; overrides in `enchantProcKindOverrides` there).
- Go side (enchants.go): one binary per (slot, candidate) with a per-slot row <= 1, so a slot may end with no pick when every candidate breaks a ceiling. The optimizer works on one clone of the request and merges the request's `player.database` before it resolves candidates. Go drops only non-static, unfit and frozen-slot candidates; the slot's input enchant does not matter, so a listed slot wearing a proc is optimized from its options (the UI decides which slots to list).
- The baseline is the input gear with a candidate slot's enchant cleared when it is a static row fitting the item; its environment build gives the base stats and the stat dependency manager. Any other enchant (Minor Speed, a proc, one with no fitting row) stays, so a slot with no pick keeps it; a slot with no pick whose input enchant is static ends empty.
- A tie-break bonus (`enchantTieBreak`; twice for the slot's input enchant, none where the input enchant stays in the baseline) keeps the input enchant on an EP tie and fills a slot whose candidates are all worth 0. The solve passes an explicit `mip_rel_gap` (`mipRelGap` in solver.go) far below HiGHS's default, which would otherwise swamp the bonus. With no candidates the optimizer returns the input gear unchanged with score 0 and 0 passes, without building an environment. The score is the objective over the baseline without the tie-break, not the gain over the input gear.
- A candidate's percent pseudo-stats are valued as `Stats.computeEP` (ui/sim/proto/stats.ts) values them: by the percent's own weight when it has one, otherwise as rating under the root rating's weight; the part all three channels of a hit, crit or haste rating share counts once at the rating, and one rating's fallbacks count once, as the largest.
- Flat weapon damage counts as main-hand, off-hand or ranged DPS at the weapon's speed.
- `StatCaps` are targets; `UndershootCaps` make them ceilings. A ceiling the baseline already exceeds allows no more of that stat.

## Architecture

Files (all under sim/core/reforge_optimizer/):

- Main flow + optimizer state: optimizer.go.
- EP scoring and objective/cap coefficient split: model.go.
- Byte-exact deterministic CPLEX LP text serialization for HiGHS: lp.go (rows are <=/>= only; term/sign formatting goes through appendLPTerm — the byte layout is load-bearing for HiGHS tie-breaking).
- Solve + cap-refinement loop: solver.go.
- Soft caps / gap-to-cap conversion + cap-detection helpers: caps.go.
- Stat/UnitStat math: reforge_stats.go.
- Enchant candidates, per-slot binaries and baseline: enchants.go.
- Gear helpers (clone, frozen slots, applying the picked enchants): gear.go.
- HiGHS bridge:
    - Go non-browser (embedded, pooled wazero): highswasm.go. (Do NOT rename to highs_wasm.go — the `_wasm.go` suffix is an implicit GOARCH=wasm constraint that breaks the native build.)
    - Browser wasm: highs_js.go.
- Frontend caller: ui/features/reforge/model/reforge_optimizer.ts (runReforgeOptimization; one slot classification shared by the skipped list and the enchant options, rebuilt when the gear, Enchanting, the frozen slots or the presets change). The IndexedDB cache key contract lives in cacheRelevantReforgeRequest (ui/sim/state/reforge_request.ts) — any new field that affects a solve must be reflected, and any irrelevant field excluded, or caches go silently stale / get needlessly busted. Hashing is synchronous via hashString in ui/sim/utils/misc.ts.

## Core Invariants

- Final correctness must be validated with exact core.ComputeStats results; solver deltas are guidance only.
- On HiGHS failure, return an error; do not silently downgrade solver behavior.
- Verbose optimizer diagnostics remain behind ReforgeOptimizeRequest.debug.
- Deterministic LP text: modelToLPFormat is byte-exact and stable, so HiGHS tie-breaking among equal-objective solutions stays reproducible.

## Cap and Breakpoint Rules

- Validate and normalize cap settings before solve.
- Enforce hard caps and breakpoint-derived limits as MIP constraints.
- If exact post-check violates constraints, tighten existing rows and re-solve.
- Keep soft-cap scoring piecewise with pre-cap and post-cap EP behavior.

## Bulk Sim Integration Contract

- The batch sim never runs the optimizer, in Go or in the UI; it crosses enchants through its enchant presets, which the optimizer's "Use enchant presets" option shares. Every optimizer request is a single-gear solve.
- Batch runs derive their RNG seed from the baseline gear and BulkSettings, with no optimizer config — see the wowsims-bulk-sim skill for the staged/finalist pipeline.

## Performance Guardrails

- Keep model-building and hot-path helpers allocation-aware.
- Keep selected-choice legality checks lightweight.
- Avoid debug timers/tracing overhead unless debug is enabled.

## Validation Commands

```bash
go test -tags with_db -count=1 ./sim/core/reforge_optimizer
go test -tags with_db -count=1 -run TestEnchantProcKind ./tools/database
npx vitest run ui/features/reforge/model/enchant_options.test.ts
npm run type-check
```

There are no preset fixtures; the optimizer's tests use the class-free `sim/core/testagent`, and the ones that build gear need the `with_db` tag. Fixtures built from a real spec's preset need every class registered, so a test that runs them belongs in `./sim`, not in the optimizer's package. Never update fixtures without being asked.

For integration changes touching bulk or endpoint behavior:

```bash
go test -tags with_db -count=1 ./sim/core/reforge_optimizer ./sim/web
```

## Fast Search Aids

```bash
rg -n "ReforgeOptimizeRequest|softCap|breakpoint|HiGHS" proto sim ui
rg -n "reforgeOptimizeAsync|ReforgeOptimize" sim ui
```

## Common Pitfalls

- Accepting solver-feasible output without exact-stat verification.
- Porting MoP-specific assumptions (reforging, Rune of Re-Origination, Windwalker/monk specs, amp trinkets).
