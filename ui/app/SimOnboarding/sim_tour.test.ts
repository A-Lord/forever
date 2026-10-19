// @vitest-environment node

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { applicableSteps } from '@features/onboarding/model/tour_steps';
import { describe, expect, it } from 'vitest';

import en from '../../../assets/locales/en/translation.json';
import { SIM_TOUR_STEPS } from './sim_tour';

const UI = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..', '..');

const productionSources = (dir: string, out: string[] = []): string[] => {
	for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
		const full = path.join(dir, entry.name);
		if (entry.isDirectory()) {
			if (entry.name !== 'generated' && entry.name !== 'testing') productionSources(full, out);
		} else if (/\.tsx?$/.test(entry.name) && !/\.(test|stories)\.tsx?$/.test(entry.name)) out.push(full);
	}
	return out;
};

describe('SIM_TOUR_STEPS', () => {
	it('has a title and a body for every step', () => {
		const steps = en.onboarding.steps as Record<string, { title?: string; body?: string }>;
		for (const step of SIM_TOUR_STEPS) {
			expect(steps[step.id]?.title, step.id).toBeTruthy();
			expect(steps[step.id]?.body, step.id).toBeTruthy();
		}
		expect(Object.keys(steps).sort()).toEqual(SIM_TOUR_STEPS.map(step => step.id).sort());
	});

	it('points only at test ids the page renders', () => {
		const source = productionSources(UI)
			.map(file => fs.readFileSync(file, 'utf8'))
			.join('\n');
		const ids = SIM_TOUR_STEPS.flatMap(step =>
			typeof step.target === 'string' ? [...step.target.matchAll(/data-testid="([^"]+)"/g)].map(match => match[1]) : [],
		);
		expect(ids.length).toBeGreaterThan(0);
		for (const id of ids) expect(source.includes(`"${id}"`) || source.includes(`'${id}'`) || source.includes(`= '${id}'`), id).toBe(true);
	});

	it('opens and closes with a centred card, and tours a gear planner without the sim', () => {
		expect(SIM_TOUR_STEPS[0].target).toBeUndefined();
		expect(SIM_TOUR_STEPS[SIM_TOUR_STEPS.length - 1].target).toBeUndefined();
		const planner = applicableSteps(SIM_TOUR_STEPS, { canSimulate: false, hasResult: false }).map(step => step.id);
		expect(planner).not.toContain('simulate');
		expect(planner).not.toContain('rotation_tab');
		expect(planner).toContain('gear');
	});
});
