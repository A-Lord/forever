import { composeStories, type Meta, setProjectAnnotations, type StoryObj } from '@storybook/react-vite';
import { storyFetch } from '@testing/storybook/fixtures/database';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import * as preview from '../../../.storybook/preview';

const annotations = setProjectAnnotations([preview]);
beforeAll(async () => {
	vi.stubGlobal('fetch', vi.fn(storyFetch));
	await annotations.beforeAll();
});
afterAll(() => vi.unstubAllGlobals());

type StoriesModule = { default: Meta } & Record<string, StoryObj>;

const modules = import.meta.glob<StoriesModule>('../../**/*.stories.tsx', { eager: true });

describe.each(Object.entries(modules))('%s', (_file, mod) => {
	const stories = Object.entries(composeStories(mod));
	it.each(stories)('%s mounts', async (_name, story) => {
		const canvasElement = document.body.appendChild(document.createElement('div'));
		try {
			await story.run({ canvasElement });
			expect(canvasElement.childElementCount).toBeGreaterThan(0);
		} finally {
			canvasElement.remove();
		}
	});
});
