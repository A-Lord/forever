import type { Spec, Stat } from '@generated/proto/common';
import { type DBSchema, type IDBPDatabase, openDB } from 'idb';

import { LOCAL_STORAGE_PREFIX } from '../constants/other';
import type { EpPrecisionSettings } from '../settings/stat_weight_settings';

const STAT_SWEEP_DB_NAME = `${LOCAL_STORAGE_PREFIX}_stat-sweep`;
const STAT_SWEEP_DB_VERSION = 1;
const RUNS_STORE = 'runs';

export interface StatSweepPoint {
	/** Bonus points of the swept stat, which % stat multipliers scale before they reach its total. */
	offset: number;
	/** The swept stat's final total with `offset` bonus points, as the sim computes it. */
	statValue: number;
	ep: number;
	/** The shown 90% ± behind `ep`. */
	epCi: number;
	dps: number;
	dpsCi: number;
	/** Paired samples behind the swept stat's weight. */
	samples: number;
}

export interface StatSweepRecord {
	id: string;
	/** Epoch milliseconds. */
	createdAt: number;
	specId: Spec;
	label: string;
	/** `UnitStat.toJson()`: IndexedDB's structured clone drops class prototypes, so a stored UnitStat would come back without its methods. */
	unitStat: object;
	refStat: Stat;
	range: number;
	step: number;
	precision: EpPrecisionSettings;
	baseStatValue: number;
	points: StatSweepPoint[];
}

interface StatSweepDb extends DBSchema {
	[RUNS_STORE]: {
		key: string;
		value: StatSweepRecord;
		indexes: { specId: number };
	};
}

// A fresh connection per call, closed when it is done, so a version upgrade in another tab is never blocked.
const openSweepDb = (): Promise<IDBPDatabase<StatSweepDb>> =>
	openDB<StatSweepDb>(STAT_SWEEP_DB_NAME, STAT_SWEEP_DB_VERSION, {
		upgrade(db) {
			const store = db.createObjectStore(RUNS_STORE, { keyPath: 'id' });
			store.createIndex('specId', 'specId');
		},
	});

export async function saveRun(run: StatSweepRecord): Promise<void> {
	const db = await openSweepDb();
	try {
		await db.put(RUNS_STORE, run);
	} finally {
		db.close();
	}
}

/** The spec's saved runs, newest first. */
export async function listRuns(specId: Spec): Promise<StatSweepRecord[]> {
	const db = await openSweepDb();
	try {
		const runs = await db.getAllFromIndex(RUNS_STORE, 'specId', specId);
		return runs.sort((a, b) => b.createdAt - a.createdAt);
	} finally {
		db.close();
	}
}

export async function deleteRun(id: string): Promise<void> {
	const db = await openSweepDb();
	try {
		await db.delete(RUNS_STORE, id);
	} finally {
		db.close();
	}
}
