import { IndividualSimSettings } from '@generated/proto/ui';
import { migrateOldProto, type ProtoConversionMap } from '@sim/proto/proto_migration';
import { updateIndividualSimProtoVersion } from '@sim/state/serialization';

// TODO: version 16 is the Forever talent rebuild -- every class's talent proto changed
// shape and StatSpellPenetration became StatSpellPiercing. No converter is registered for
// it on purpose: a saved TBC build's talent string has no meaning against a Forever tree,
// so there is nothing to migrate it to. Settings below version 16 fall through and the
// user re-picks their talents.
const TBC_CONVERSION_MAP: ProtoConversionMap<IndividualSimSettings> = new Map([
	// Version 15 renamed the shadow priest's Player oneof from `priest` to `dps_priest` (and the
	// Spec value to SpecDpsPriest): the sim covers any dps priest, not only shadow. Binary payloads
	// (share links) carry the field by number, and JSON ones (autosaved settings, the JSON importer)
	// are renamed before parsing by `parseLegacySettingsJson`, so there is nothing left to convert
	// here; the entry only stamps the version.
	[15, (oldProto: IndividualSimSettings) => oldProto],
]);

/**
 * TBC's own migrations, then the shared ones. The order matters: the shared pass stamps the proto
 * as current, after which nothing below would ever run again.
 */
export function updateIndividualProtoVersion(settingsProto: IndividualSimSettings) {
	migrateOldProto(settingsProto, settingsProto.apiVersion, TBC_CONVERSION_MAP);
	updateIndividualSimProtoVersion(settingsProto);
}
