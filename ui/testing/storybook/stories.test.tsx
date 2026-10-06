import { composeStories, type Meta, setProjectAnnotations, type StoryObj } from '@storybook/react-vite';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import * as preview from '../../../.storybook/preview';

const annotations = setProjectAnnotations([preview]);
beforeAll(async () => {
	await annotations.beforeAll();
});

type StoriesModule = { default: Meta } & Record<string, StoryObj>;

const modules = import.meta.glob<StoriesModule>('../../**/*.stories.tsx', { eager: true });

describe.each(Object.entries(modules))('%s', (_file, mod) => {
	const stories = Object.entries(composeStories(mod));
	it.each(stories)('%s mounts', async (_name, story) => {
		const canvasElement = document.body.appendChild(document.createElement('div'));
		const fetchCalls = vi.spyOn(globalThis, 'fetch');
		try {
			await story.run({ canvasElement });
			expect(canvasElement.childElementCount).toBeGreaterThan(0);
			const databaseRequests = fetchCalls.mock.calls.filter(([input]) => {
				const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
				return new URL(url, location.href).pathname.includes('/assets/database/');
			});
			expect(databaseRequests).toEqual([]);
		} finally {
			fetchCalls.mockRestore();
			canvasElement.remove();
		}
	});
});
