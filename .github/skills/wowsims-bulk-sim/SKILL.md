---
name: wowsims-bulk-sim
description: 'Use when working on WoWSims Forever Bulk Sim: local/server bulk sim, browser WASM concurrent bulk sim, BulkSimRequest/BulkSimResult protos, candidate generation, staged simulation, the finalist tie-breaker stage, deterministic seeds, tie-group result display, progress, or abort behavior.'
argument-hint: 'Describe the Bulk Sim candidate flow, staging/finalist behavior, or validation task.'
---

# WoWSims Forever Bulk Sim Guide

Comments, docs, commit messages and PR text follow `wowsims-conventions`: "game files", not "client", no history in comments, no "golden".

## Scope

- Bulk candidate generation and staged simulation flow (low/medium/high + finalist).
- There is no optimizer pre-pass: generated candidates go straight to the staged sim ("reforge" is only the gear optimizer's name; there is no reforging).
- Bulk crosses enchants through its enchant presets. They live in ui/features/enchant-presets/ (model, `EnchantPresets` tiles and dialog) and are shared with the gear optimizer's "Use enchant presets" option: same storage (`bulk-enchant-presets.v1` through ui/sim/settings/bulk_settings.ts, the bulk slice's `enchantPresets`), same checks. Checking a preset, or saving or deleting a checked one, bumps the bulk `settings` counter.
- Bulk's enchant fit rule lives in sim/core (`core.EnchantAppliesToItem`, `core.FittingEnchantRow`), shared with the gear optimizer.
- Deterministic content-derived seeding and statistical result display.
- Progress, abort semantics, and local vs WASM orchestration.

## Architecture

- Shared messages: proto/api.proto (BulkSimRequest, BulkSimResult, BulkGearResult). BulkSimStage order: Low, Medium, High, Finalist, Complete (Complete last).
- Candidate generator: sim/core/bulk/generator.go (options, pairs, combination table), weapons.go (main-hand/off-hand combos), enchants.go (enchant pools), candidates.go (entry points).
- Core staged runner: sim/core/bulk/bulk_sim.go; per-stage logic + finalist stage: sim/core/bulk/stage.go; paired statistics: sim/core/bulk/statistics.go.
- Local/server entry: sim/web/bulk.go (generates the candidates when the request carries none, then hands over to sim/core/bulk); endpoint wiring: sim/web/main.go.
- Frontend orchestration, content seed: ui/sim/sim.ts (runBulkSim) and ui/features/bulk/model/run.ts (runBulkBatch).
- Bulk utilities: ui/sim/bulk/utils.ts.
- Results display (tie groups, ±CI, change icons): ui/features/bulk/components/BulkResults/ + ui/features/bulk/model/tie_chains.ts + ui/features/gear/components/GearChangeIcon/ (Wowhead cover link; no reforge glyph).
- Generic cache storage: ui/sim/cache/reforge_cache.ts (sync hashString from ui/sim/utils/misc.ts; values are plain `equipmentSpec:`-prefixed proto JSON; age-index-ranged prune).
- Browser WASM path: ui/sim/wasm/bulk_sim/ (index, stage, batch, merge, statistics, progress, carry_over, estimate, types) — a deliberate line-for-line mirror of sim/core/bulk.
- Generated constants (never hand-edit; `go run ./tools/database/gen_db -gen=go-to-ts`): ui/sim/bulk/constants_auto_gen.ts (slot maps) and ui/sim/wasm/bulk_sim/constants_auto_gen.ts (tuning constants + stage ladder), emitted by tools/database/gen_bulksim_constants.ts.go from the exported consts in sim/core/bulk.

## Game Rules the Generator Encodes

- No reforging and no gems or sockets: ItemSpec carries id, random suffix and enchant only; no reforge glyph anywhere in the UI.
- No Titan's Grip: a two-hander never occupies the off-hand.
- The Ranged slot is a real bulk slot that iterates like any non-weapon slot (bulkSimNonWeaponOrder is built by filtering, not slicing).
- No required-set-bonus combination matcher (matchingCombinations == rawCombinations).
- A ring enchant needs Enchanting: for a player without it (Profession1/Profession2 on the player proto), no candidate wears a ring enchant. The generator drops the ring enchant pool and strips both ring slots (`slotSpec`), the baseline's and a selected ring's own enchant included, so no two candidates differ only by a ring enchant.

## Determinism and Statistics

- Content-derived seed: runBulkSim hashes baseline gear + BulkSettings into simOptions.randomSeed (utils hashString; per-part digests combined). Same setup → bit-identical results; any change → fresh sample. An explicit fixed RNG seed takes precedence; lastUsedRngSeed is updated to the seed actually used.
- Candidates share seed sequences; comparisons between candidates use PAIRED errors (bulkSimPairedDpsError over AllValues), far tighter than per-result stdev.
- Finalist stage (runBulkSimFinalistStage / runConcurrentBulkSimFinalistStage): after the high stage, the top `topResults` candidates + baseline get lockstep extra iterations until every adjacent pair separates under a paired z-test at Z95 (bulkSimZ95 / Z_95), or the budget (BulkSimFinalistMaxExtraIterationMultiplier × high-stage iterations) is spent. Its returned results ARE the refined, DPS-sorted display set.
- Shipped per top result: paired_error_to_next_result and paired_error_to_baseline (computed before AllValues are stripped; 0 = could not pair).
- FE display: adjacent results inside the paired tie threshold render in one labeled tie group (unpaired zTest fallback when paired data is absent); each row shows a ±95% CI (stDevToConf95 — the unpaired per-row error, a different quantity from the pairwise grouping test).
- The single significance threshold is BulkSimZ95 in sim/core/bulk/statistics.go; `make go-to-ts` writes it to the UI as BULK_SIM_Z95, which Z_95 and zTest in ui/sim/utils/math.ts read.

## Core Invariants

- Baseline gear source is base_request.raid.parties[0].players[0].equipment.
- Candidate identity remains stable through BulkGearCandidate.index.
- Dedup for sim input must exclude baseline-equivalent and duplicate gear (`dedupeGearSets` over EquipmentSpecs, keyed by `getGearIdentityKey`).
- Spec lookup goes through core.PlayerProtoToSpecSafe; eligible-slot logic through core.EligibleSlotsForItem / core.ItemTypeToSlotsMap (single source, shared with item swap); entry-point validation via newGeneratorFromRequest.

## Settings Boundaries

- BulkSettings controls bulk-tab constraints.
- ReforgeSettings belongs to the gear optimizer; the batch sim reads none of it. The two share only the enchant presets, which live in the bulk slice.
- Never merge or alias these domains in request shaping or serialization.

## Local/Server Flow

- A request with no candidates gets them from the generator in-process (`bulk.EnsureBulkSimCandidatesGenerated` in sim/web/bulk.go), then runs the low/medium/high stages.
- Nothing is emitted while the candidates are built. The combination count comes from BulkCombinationCount before the run; during the stages, progress carries TotalSims, and the final result's StageMetrics carry InputGearSets per stage.

## Frontend/WASM Flow

- The WASM path makes no per-gear optimizer calls and neither reads nor writes the reforge cache.
- Candidates stay EquipmentSpecs from the generator to the workers: native runs send none and Go builds them in-process; on WASM, Sim.runBulkSim gets them through getBulkCandidates and dedupes them against the prepared baseline. The generator strips a non-enchanter's ring enchants on both paths, the per-candidate request adjusts imbues from the base request's consumables and the weapon items, and Gear is built only for the displayed top results.
- The request database comes from the baseline, BulkSettings.items (a selected spec's own enchant and suffix included) and the enchant pool (`makeBulkItemDatabaseFromSpecs`) on both paths, never from the candidates; an id a candidate can carry reaches the workers only through BulkSettings.
- Results tab: rows build into a DocumentFragment and attach once; changed items render the shared gear-change icon (Wowhead cover link) instead of a border; a slot-change renders the item plainly. Starting a new run returns to the Setup tab.

## Candidate Counting and Filtering

- rawCombinations is the mixed-radix candidate index space.
- combinations is the filtered runnable count.
- User-visible progress/counts should reflect filtered candidate totals.
- Hand swap (sim/core/bulk/generator.go initSelectedItems): with no weapon among the selected items, the worn weapons go into the MainHand and OffHand bulk slots and stay in the hands they are worn in; once any weapon is selected, a dual-wielder's worn weapons join the HandWeapon pool and every allowed main-hand/off-hand order is a candidate.

## Performance Guardrails

- Avoid per-candidate allocations in hot loops.
- Candidate generation (sim/core/bulk/generator.go combinationTable) resolves every choice's item once per physical slot; a candidate copies the baseline item specs, overwrites the chosen slots and emits its 17 ItemSpecs from one backing array (4 allocations per candidate).
- Prefer preallocated imperative loops in candidate/cache helpers.
- Progress emission is throttled (BulkSimProgressThrottle Go-side, 100ms mirror in wasm progress emitter).
- When stripping AllValues for proto output, detach the slice before cloning (see bulkSimCandidateResultToProto) — never clone megabytes just to discard them.

## Logging Expectations

- Candidate generation in sim/web logs its completion with duration and candidate count.
- Every stage (finalist included) reports metrics through the shared stage-metrics builder, so observed error and concurrency are always populated.

## Validation Commands

```bash
make proto
npm run type-check
go test -tags with_db -count=1 ./sim/core/bulk ./sim/web
```

Reproducibility check: run the same BulkSimRequest twice through BulkSimAsync with the same seed — results must be bit-identical.

## Fast Search Aids

```bash
rg -n "Finalist" proto sim ui
rg -n "bulkSimAsync|/bulkSimAsync" sim/web
rg -n "pairedErrorToNextResult|bulkSimPairedDpsError|Z_95|BulkSimZ95|BULK_SIM_Z95" sim ui
```

## Common Pitfalls

- Conflating BulkSettings and ReforgeSettings.
- Porting MoP-only assumptions into request/proto shapes (reforging, Titan's Grip, required set bonuses, a second locale).
