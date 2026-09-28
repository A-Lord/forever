export interface ThemeVar {
	name: string; // --color-primary
	value: string; // raw source text, display only
	inline: boolean; // declared in a `@theme static inline` block
	prefix: string; // color | spacing | text | shadow | …
}

const BLOCK_RE = /@theme\s+static(\s+inline)?\s*\{([\s\S]*?)\n\}/g;
export const VAR_RE = /^\s*(--[a-z0-9-]+):\s*([\s\S]+?);\s*$/gm;

export const parseThemeVars = (css: string): ThemeVar[] => {
	const out: ThemeVar[] = [];
	for (const block of css.matchAll(BLOCK_RE)) {
		const inline = !!block[1];
		for (const m of block[2].matchAll(VAR_RE)) {
			if (m[1].endsWith('-*')) continue;
			out.push({ name: m[1], value: m[2].replace(/\s+/g, ' '), inline, prefix: m[1].slice(2).split('-')[0] });
		}
	}
	return out;
};
