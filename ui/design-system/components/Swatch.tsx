import type { ThemeVar } from '../tokens/parse_theme';

interface SwatchProps {
	token: ThemeVar;
	foreground?: ThemeVar;
}

export const Swatch = ({ token, foreground }: SwatchProps) => (
	<div className="flex items-center gap-3 border border-border-muted p-2" data-testid="ds-swatch">
		<span
			className="flex size-10 shrink-0 items-center justify-center border border-border font-bold"
			style={{ backgroundColor: `var(${token.name})`, color: foreground && `var(${foreground.name})` }}>
			{foreground && 'Aa'}
		</span>
		<div className="min-w-0 flex-1">
			<code className="text-sm">{token.name}</code>
			{token.inline && <span className="ml-2 text-ui text-warning">inline</span>}
			<div className="truncate font-mono text-ui text-link" title={token.value}>
				{token.value}
			</div>
			{foreground && <code className="block truncate text-ui">{foreground.name}</code>}
		</div>
	</div>
);
