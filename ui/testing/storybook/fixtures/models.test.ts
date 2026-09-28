// @vitest-environment node
import { existsSync, readFileSync, statSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

const UI = resolve(dirname(fileURLToPath(import.meta.url)), '../../..');
const ALIASES = ['sim', 'ui-kit', 'features', 'generated', 'i18n'];

// Runtime imports only: `import type` never reaches the browser, and a bare package is out of scope.
const IMPORT = /(?:import|export)\s+(?!type\s)(?:[^'";]*?\sfrom\s+)?['"]([^'"]+)['"]|import\(\s*['"]([^'"]+)['"]\s*\)/g;

const toFile = (base: string) =>
	[base, `${base}.ts`, `${base}.tsx`, `${base}/index.ts`, `${base}/index.tsx`].find(path => existsSync(path) && statSync(path).isFile());

const resolveSpecifier = (from: string, specifier: string): string | undefined => {
	if (specifier.startsWith('.')) return toFile(resolve(dirname(from), specifier));
	const alias = ALIASES.find(name => specifier.startsWith(`@${name}/`));
	return alias ? toFile(resolve(UI, alias, specifier.slice(alias.length + 2))) : undefined;
};

const runtimeGraph = (entry: string) => {
	const importedBy = new Map<string, string>([[entry, entry]]);
	const queue = [entry];
	for (let file = queue.shift(); file; file = queue.shift()) {
		for (const match of readFileSync(file, 'utf8').matchAll(IMPORT)) {
			const next = resolveSpecifier(file, match[1] ?? match[2]);
			if (next && /\.tsx?$/.test(next) && !importedBy.has(next)) {
				importedBy.set(next, file);
				queue.push(next);
			}
		}
	}
	return importedBy;
};

describe('story fixtures', () => {
	it('never load a module that imports vitest, which throws in the Storybook dev server', () => {
		const graph = runtimeGraph(resolve(UI, 'testing/storybook/fixtures/models.tsx'));
		const offenders = [...graph.keys()]
			.filter(file => /from ['"]vitest['"]/.test(readFileSync(file, 'utf8')))
			.map(file => `${file.slice(UI.length + 1)} (imported by ${graph.get(file)!.slice(UI.length + 1)})`);

		expect(graph.size).toBeGreaterThan(10);
		expect(offenders).toEqual([]);
	});
});
