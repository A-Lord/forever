// Settings JSON written before a proto rename. Share links are binary, so field numbers carry
// them across; only the JSON forms (autosaved settings in localStorage and the JSON importer) name
// fields and need a hand.
import { migrateLegacyConsumes } from '../proto/consumes_migration';

// The shadow priest's Player oneof was `priest` until the spec became `dps_priest` (the sim covers
// any dps priest, not only shadow). Renames the field in place so an old blob parses.
const migrateLegacyPriest = (json: unknown): unknown => {
	if (!json || typeof json !== 'object') return json;
	const player = (json as { player?: unknown }).player;
	if (!player || typeof player !== 'object') return json;
	const spec = player as Record<string, unknown>;
	if ('priest' in spec && !('dpsPriest' in spec)) {
		const { priest, ...rest } = spec;
		return { ...(json as object), player: { ...rest, dpsPriest: priest } };
	}
	return json;
};

export const migrateLegacySettingsJson = (json: unknown): unknown => migrateLegacyConsumes(migrateLegacyPriest(json));

// JSON text in, migrated object out; a parse failure surfaces to the caller as before.
export const parseLegacySettingsJson = (text: string): unknown => migrateLegacySettingsJson(JSON.parse(text));
