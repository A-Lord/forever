# Persistence and serialization

**Source of truth:** `ui/sim/state/persistence.ts` (load order + autosave),
`ui/sim/state/serialization.ts` (the `IndividualSimSettings` envelope),
`ui/sim/state/sim_links.ts` (URL-hash import), `ui/sim/state/stat_filter_sharing.ts` (stat filters
stay with the viewer). The load-order contract is written as a comment at
the top of `persistence.ts` — read it there if this file and the code disagree:

```
sed -n '1,20p' ui/sim/state/persistence.ts
```

All four are browser-free: the localStorage and `location` surfaces arrive through the injected
`Env` (`sim.env`), never through globals.

## The load-order contract — do not reorder

`loadIndividualSettings(host, opts)` runs the whole sequence inside one `batch()`:

1. `host.applyDefaults()`
2. saved settings from `env.storage` (parse failures are warned and swallowed — a corrupt blob must
   not brick the page). Read through `fromJsonStringLenient` (`ui/sim/proto/lenient_json.ts`): a plain
   `ignoreUnknownFields` read leaves a oneof holding an enum name the protos do not list with an
   undefined value, and the next `toBinary` throws
3. the URL-hash link, via `tryParseUrlLocation(env.location)`. **After** step 2 on purpose: a
   partial link import (rotation only, say) must keep the previous settings for the other categories.
   `withViewerStatFilters` writes the sim's stat filters into the link's filters before it applies,
   so the saved ones stay, whatever the link carries
4. `env.location.setHash('')`
5. `player.setName('Player')`
6. `opts.autosaveSubscribe(schedulePersist)` — **last**, so initialization does not re-store what it
   just loaded
7. `opts.statWeightSettings.load()`

Then one explicit `persist()` outside the batch, because the subscription registered in step 6 only
sees changes made after it existed.

Autosave is debounced by `AUTOSAVE_DEBOUNCE_MS` (300 ms): serializing and storing the full envelope
on every keystroke cost ~50 ms per APL edit. A pending write is flushed by `env.onPageHide`.

**The trap this cost an afternoon:** `persistTimer` must be declared _before_ the load batch. The
batch's flush can already schedule a persist, so declaring the timer after it is a TDZ error that
only fires on a path with saved settings.

## Storage keys

The key is built by the caller and handed in — the state layer never derives one.
`SimHostObject.getStorageKey(postfix)` (`ui/app/individual_sim_ui.tsx`) is the builder, over
`specStorageKey` in `ui/sim/state/storage_keys.ts`; `ui/sim/constants/other.ts` holds
`LOCAL_STORAGE_PREFIX`. Suffixes in use — the saved-data ones live in `storage_keys.ts`, the first
two in `persistence.ts`:

| Constant                                                  | Key                                                                                            |
| --------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `SETTINGS_STORAGE_SUFFIX`                                 | the settings envelope, per spec                                                                |
| `SHARED_SAVED_ENCOUNTER_STORAGE_KEY`                      | saved encounters — deliberately **not** per-spec prefixed, so they are shared across every sim |
| `STAT_WEIGHT_SETTINGS_STORAGE_KEY`                        | stat-weight action settings (`'__statweight_settings__'`)                                      |
| `SAVED_GEAR_STORAGE_KEY` / `SAVED_EP_WEIGHTS_STORAGE_KEY` | the saved-data slots                                                                           |

If you add a key, add it via `getStorageKey` unless you positively want it shared across specs, and
say which in a comment — the saved-encounter exception is the only one today and it is easy to
imitate by accident.

## The envelope

`individualSimSettingsToProto(ctx, exportCategories?)` and `applyIndividualSimSettings(...)` in
`serialization.ts` assemble and apply `IndividualSimSettings`. `IndividualSimUI.toProto` /
`fromProto` are thin wrappers. The envelope covers more than the store:
`IndividualSimSerializationContext` also carries the reforge settings model and the EP
reference-stat selections owned by the sim UI.

`updateIndividualSimProtoVersion` migrates older blobs — a settings shape change usually belongs
there rather than in a new field default.

Category filtering is shared with link import: `SimSettingCategories`, and
`LINK_DEFAULT_CATEGORIES` in `sim_links.ts` is every category except `UISettings` (your show-flags
and language do not travel through a shared link).

The stat filters (`DatabaseFilters.statFilters`) belong to the viewer and never travel. The link
and JSON exporters (`ui/features/import-export/exporters/link.ts`, `ui/features/import-export/exporters/json.ts`) strip them
with `withoutStatFilters`, whatever the categories, and the JSON importer
(`ui/features/import-export/importers/json.ts`) writes the viewer's into the envelope with
`withViewerStatFilters` before `fromProto`, as the link import in step 3 does. Both helpers change
the envelope they are given and return it. An envelope without filters needs no patch: loading it
falls back to `Sim.defaultFilters()` (or `applyDefaults` without settings), which keeps the viewer's
stat filters. The autosaved envelope keeps them.

## What is gated

No suite pins the serialized shape. Tests do run through `serialization.ts`: the
`makeIndividualSim` fixture (`ui/testing/sim/individual_sim.ts`) wires `toProto` / `fromProto` to
it, and the persistence, import-export and shared-build tests use that fixture. They check what a
round trip keeps, not the envelope as a whole, so a persistence change that alters the envelope can
still pass every gate. Check it by hand — serialize a default page before and after the change and
diff the two.

Two `Sim` round-tripping quirks will bite you here specifically:

- `Sim.toProto` collapses all-selected filter arrays to `[]`; `Sim.fromProto` re-expands them, so
  the form is only a fixed point from the second pass. Canonicalize once before comparing.
- `Sim.fromProto` mutates its **argument** in place. Serialize before round-tripping.
