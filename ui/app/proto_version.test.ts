import { Player } from '@generated/proto/api';
import { ConsumesSpec, PartyBuffs } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import { CURRENT_API_VERSION } from '@sim/constants/other';
import { describe, expect, it } from 'vitest';

import { updateIndividualProtoVersion } from './proto_version';

const settings = (apiVersion: number) =>
	IndividualSimSettings.create({
		apiVersion,
		player: Player.create({ consumables: ConsumesSpec.create({ potId: 13444 }) }),
		partyBuffs: PartyBuffs.create(),
	});

describe('updateIndividualProtoVersion', () => {
	it('stamps every migrated payload as current, so it is not migrated twice', () => {
		const proto = settings(6);

		updateIndividualProtoVersion(proto);

		expect(proto.apiVersion).toBe(CURRENT_API_VERSION);
	});

	it('stamps a version-14 payload as current without touching it: the priest oneof rename happens before parsing', () => {
		const proto = settings(14);

		updateIndividualProtoVersion(proto);

		expect(proto.apiVersion).toBe(CURRENT_API_VERSION);
		expect(proto.player?.consumables?.potId).toBe(13444);
	});
});
