import type { IndividualSimHost } from './sim_host';
import type { StoreSubscribe } from './state/subscriptions';

// No vitest import here or below: Storybook loads this module in the browser through the story fixtures.

export const noopSubscribe: StoreSubscribe = () => () => {};

export type FakeHostParts = Partial<Record<keyof IndividualSimHost<any>, unknown>>;

export const fakeHost = (parts: FakeHostParts = {}): IndividualSimHost<any> =>
	({
		player: {},
		sim: {},
		disabled: false,
		simDisabled: false,
		config: { className: '', cssScheme: '' },
		individualConfig: {},
		...parts,
	}) as unknown as IndividualSimHost<any>;
