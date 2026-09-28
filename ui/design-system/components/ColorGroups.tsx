import { Unstyled } from '@storybook/addon-docs/blocks';

import colorsCss from '../../styles/theme/colors.css?raw';
import { parseThemeVars, type ThemeVar } from '../tokens/parse_theme';
import { Swatch } from './Swatch';

const GROUPS = ['base', 'class', 'school', 'quality', 'resource', 'damage', 'timeline', 'faction', 'talent', 'cr'] as const;
type Group = (typeof GROUPS)[number];

const groupOf = (token: ThemeVar): Group => {
	const segment = token.name.split('-')[3];
	return GROUPS.find(group => group !== 'base' && group === segment) ?? 'base';
};

const colors = parseThemeVars(colorsCss).filter(token => token.prefix === 'color');
const byName = new Map(colors.map(token => [token.name, token]));

export const ColorGroups = () => (
	<Unstyled>
		{GROUPS.map(group => {
			const tokens = colors.filter(token => groupOf(token) === group && !(group === 'class' && token.name.endsWith('-foreground')));
			return (
				<section key={group}>
					<h2>
						{group} ({tokens.length})
					</h2>
					<div className="grid grid-cols-auto-fit-220 gap-2">
						{tokens.map(token => (
							<Swatch key={token.name} token={token} foreground={group === 'class' ? byName.get(`${token.name}-foreground`) : undefined} />
						))}
					</div>
				</section>
			);
		})}
	</Unstyled>
);
