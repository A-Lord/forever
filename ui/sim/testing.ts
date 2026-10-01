import { vi } from 'vitest';

import { noopSubscribe } from './fake_host';
import type { StoreSubscribe } from './state/subscriptions';

export { fakeHost, type FakeHostParts, noopSubscribe } from './fake_host';

type Subscriptions = typeof import('./state/subscriptions');

/**
 * Every `subscribe*` export answered from one source; non-subscription exports stay real.
 *
 * `vi.mock` is hoisted above the imports, so the factory has to reach this module lazily:
 *
 * ```ts
 * vi.mock('@sim/state/subscriptions', async () => (await import('@sim/testing')).mockSubscriptions(source.subscribe));
 * ```
 */
export const mockSubscriptions = async (
	source: StoreSubscribe = noopSubscribe,
	overrides: Partial<Record<keyof Subscriptions, unknown>> = {},
): Promise<Subscriptions> => {
	const actual = await vi.importActual<Subscriptions>('./state/subscriptions');
	const answered = Object.keys(actual).map(name => [name, name.startsWith('subscribe') ? () => source : (actual as Record<string, unknown>)[name]]);
	return { ...Object.fromEntries(answered), ...overrides } as Subscriptions;
};
