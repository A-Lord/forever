import { StatFilterPopover } from '@features/stat-filters';
import { Class, PseudoStat, Stat } from '@generated/proto/common';
import { UnitStat } from '@sim/proto/stats';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { Table, TableCell, TableRow } from '@ui-kit/Table';
import { fn } from 'storybook/test';

import { CharacterStatsHeaderView, CharacterStatsPanelView } from './CharacterStats';

const ROWS = [
	['Strength', '320'],
	['Agility', '180'],
	['Stamina', '410'],
	['Attack Power', '1,240'],
	['Melee Crit', '24.50%'],
];

const SampleTable = () => (
	<Table variant="plain" flush>
		<tbody>
			{ROWS.map(([label, value]) => (
				<TableRow key={label} className="ui-character-stats-row">
					<TableCell className="ui-character-stats-label">{label}</TableCell>
					<TableCell className="ui-character-stats-value">{value}</TableCell>
				</TableRow>
			))}
		</tbody>
	</Table>
);

const UNIVERSE = [
	UnitStat.fromStat(Stat.StatStrength),
	UnitStat.fromStat(Stat.StatAgility),
	UnitStat.fromStat(Stat.StatStamina),
	UnitStat.fromStat(Stat.StatAttackPower),
	UnitStat.fromPseudoStat(PseudoStat.PseudoStatMeleeCritPercent),
	UnitStat.fromStat(Stat.StatSpellDamage),
];
const SPEC_DEFAULTS = UNIVERSE.filter(stat => !stat.equalsStat(Stat.StatSpellDamage));

const statFilter = (customised: boolean) => (
	<StatFilterPopover
		universe={UNIVERSE}
		shown={customised ? UNIVERSE : SPEC_DEFAULTS}
		customised={customised}
		playerClass={Class.ClassWarrior}
		onChange={fn()}
		onReset={fn()}
		testId="character-stats-filter"
	/>
);

const meta = {
	title: 'Features/CharacterStats/CharacterStats',
	component: CharacterStatsPanelView,
	subcomponents: { CharacterStatsHeaderView },
	tags: ['autodocs'],
	args: { title: 'Stats', filter: statFilter(false), children: <SampleTable /> },
	argTypes: { filter: { control: false }, children: { control: false } },
	decorators: [
		Story => (
			<div className="flex w-62 flex-col">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof CharacterStatsPanelView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const CustomisedFilter: Story = { args: { filter: statFilter(true) } };

export const Header: Story = { render: args => <CharacterStatsHeaderView {...args} /> };
