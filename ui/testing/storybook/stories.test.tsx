import { readFileSync } from 'node:fs';

import { UIDatabase } from '@generated/proto/ui';
import { composeStories, type Meta, setProjectAnnotations, type StoryObj } from '@storybook/react-vite';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import * as preview from '../../../.storybook/preview';

const DATABASE_URL = '/forever/assets/database/db.json';
const databaseJson = readFileSync('assets/database/db.json', 'utf8');
const storyDatabase = UIDatabase.fromJson(JSON.parse(databaseJson));
const storyFetch = vi.fn((url: RequestInfo | URL) => {
	const requested = String(url);
	if (requested === DATABASE_URL) return Promise.resolve({ json: () => Promise.resolve(JSON.parse(databaseJson)) });
	const match = new URL(requested).pathname.match(/^\/forever\/tooltip\/(item|spell)\/(\d+)$/);
	if (match?.[1] === 'item') {
		const item = storyDatabase.items.find(entry => entry.id === Number(match[2]));
		return Promise.resolve({ json: () => Promise.resolve({ name: item?.name ?? '', icon: item?.icon ?? '', buff: '' }) });
	}
	if (match?.[1] === 'spell') {
		const spell = storyDatabase.spellIcons.find(entry => entry.id === Number(match[2]));
		return Promise.resolve({ json: () => Promise.resolve({ name: spell?.name ?? '', icon: spell?.icon ?? '', buff: spell?.hasBuff ? '1' : '' }) });
	}
	return Promise.reject(new Error(`unexpected Storybook fetch: ${String(url)}`));
});

const annotations = setProjectAnnotations([preview]);
beforeAll(async () => {
	vi.stubGlobal('fetch', storyFetch);
	await annotations.beforeAll();
});
afterAll(() => vi.unstubAllGlobals());

type StoriesModule = { default: Meta } & Record<string, StoryObj>;

const modules = import.meta.glob<StoriesModule>('../../**/*.stories.tsx', { eager: true });

describe.each(Object.entries(modules))('%s', (_file, mod) => {
	const stories = Object.entries(composeStories(mod));
	it.each(stories)('%s mounts', async (_name, story) => {
		if (story.parameters?.smoke === false) {
			expect(story.parameters.smokeReason).toMatch(/\S/);
			return;
		}
		const canvasElement = document.body.appendChild(document.createElement('div'));
		try {
			await story.run({ canvasElement });
			expect(canvasElement.childElementCount).toBeGreaterThan(0);
		} finally {
			canvasElement.remove();
		}
	});
});
