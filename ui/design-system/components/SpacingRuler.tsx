import { Unstyled } from '@storybook/addon-docs/blocks';

import breakpointsCss from '../../styles/theme/breakpoints.css?raw';
import { parseRootVars, THEME_TOKENS } from '../tokens/theme_tokens';

const PLAIN_LENGTH = /^[\d.]+(px|rem)$/;

export const SpacingRuler = () => (
	<Unstyled>
		<div className="mb-6 flex flex-col gap-1">
			{THEME_TOKENS.filter(token => ['spacing', 'size', 'container'].includes(token.prefix) && PLAIN_LENGTH.test(token.value)).map(token => (
				<div key={token.name} className="flex items-center gap-4">
					<code className="w-64 shrink-0 text-ui">
						{token.name}: {token.value}
					</code>
					<span className="block h-3 max-w-full bg-primary" style={{ width: `var(${token.name})` }} />
				</div>
			))}
		</div>
	</Unstyled>
);

const VARIANT_RE = /@variant\s+([a-z-]+)\s*\{([^}]*)\}/g;

export const ResponsiveSpacing = () => (
	<Unstyled>
		<table className="mb-6 w-full text-sm">
			<thead>
				<tr className="border-b border-border">
					<th className="p-2">Variant</th>
					<th className="p-2">Variable</th>
					<th className="p-2">Value</th>
				</tr>
			</thead>
			<tbody>
				{[...breakpointsCss.matchAll(VARIANT_RE)].flatMap(([, variant, body]) =>
					parseRootVars(body).map(token => (
						<tr key={`${variant}${token.name}`} className="border-b border-border-muted">
							<td className="p-2">
								<code>{variant}</code>
							</td>
							<td className="p-2">
								<code>{token.name}</code>
							</td>
							<td className="p-2 font-mono text-link">{token.value}</td>
						</tr>
					)),
				)}
			</tbody>
		</table>
	</Unstyled>
);
