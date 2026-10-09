import { IndividualSimSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Database } from '@sim/proto/database';
import { batch } from '@sim/state/batch';
import { restoreViewerStatFilters } from '@sim/state/stat_filter_sharing';

import type { ImporterDefinition } from './types';

export const JSON_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.json.title'),
	allowFileUpload: true,
	onImport: async (host, data) => {
		let proto: ReturnType<typeof IndividualSimSettings.fromJson>;
		try {
			proto = IndividualSimSettings.fromJsonString(data, { ignoreUnknownFields: true });
		} catch {
			throw new Error(i18n.t('import.json.error_invalid_json'));
		}
		if (proto.player?.equipment) {
			await Database.loadLeftoversIfNecessary(proto.player.equipment);
		}
		const statFilters = host.sim.getFilters().statFilters;
		batch(() => {
			host.fromProto(proto);
			restoreViewerStatFilters(host.sim, statFilters);
		});
	},
};
