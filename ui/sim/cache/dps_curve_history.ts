import type { Spec } from '@generated/proto/common';
import { type DBSchema, type IDBPDatabase, openDB } from 'idb';

import { LOCAL_STORAGE_PREFIX } from '../constants/other';
import type { DpsCurvePrecision } from '../settings/dps_curve_settings';

const DPS_CURVE_DB_NAME = `${LOCAL_STORAGE_PREFIX}_dps-curve`;
const DPS_CURVE_DB_VERSION = 1;
const RUNS_STORE = 'runs';

export interface DpsCurvePoint {
	/** Bonus points of the swept stat, which % stat multipliers scale before they reach its total. */
	offset: number;
	/** The swept stat's final total with `offset` bonus points, as the sim computes it. */
	statValue: number;
	dps: number;
	dpsStdev: number;
	/** Iterations behind `dps`. */
	samples: number;
}

export interface DpsCurveRecord {
	id: string;
	/** Epoch milliseconds. */
	createdAt: number;
	specId: Spec;
	label: string;
	/** The chart colour slot the run keeps across reloads. */
	colorSlot: number;
	/** `UnitStat.toJson()`: IndexedDB's structured clone drops class prototypes, so a stored UnitStat would come back without its methods. */
	unitStat: object;
	range: number;
	step: number;
	precision: DpsCurvePrecision;
	baseStatValue: number;
	points: DpsCurvePoint[];
}

interface DpsCurveDb extends DBSchema {
	[RUNS_STORE]: {
		key: string;
		value: DpsCurveRecord;
		indexes: { specId: number };
	};
}

// A fresh connection per call, closed when it is done, so a version upgrade in another tab is never blocked.
const openDpsCurveDb = (): Promise<IDBPDatabase<DpsCurveDb>> =>
	openDB<DpsCurveDb>(DPS_CURVE_DB_NAME, DPS_CURVE_DB_VERSION, {
		upgrade(db) {
			const store = db.createObjectStore(RUNS_STORE, { keyPath: 'id' });
			store.createIndex('specId', 'specId');
		},
	});

export async function saveRun(run: DpsCurveRecord): Promise<void> {
	const db = await openDpsCurveDb();
	try {
		await db.put(RUNS_STORE, run);
	} finally {
		db.close();
	}
}

/** The spec's saved runs, newest first. */
export async function listRuns(specId: Spec): Promise<DpsCurveRecord[]> {
	const db = await openDpsCurveDb();
	try {
		const runs = await db.getAllFromIndex(RUNS_STORE, 'specId', specId);
		return runs.sort((a, b) => b.createdAt - a.createdAt);
	} finally {
		db.close();
	}
}

export async function deleteRun(id: string): Promise<void> {
	const db = await openDpsCurveDb();
	try {
		await db.delete(RUNS_STORE, id);
	} finally {
		db.close();
	}
}
