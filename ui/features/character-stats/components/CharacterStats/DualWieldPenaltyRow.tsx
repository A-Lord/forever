import i18n from '@i18n/config';

export const DualWieldPenaltyRow = () => (
	<tr data-testid="character-stats-table-row" className="ui-character-stats-row">
		<td colSpan={2} className="ui-character-stats-label text-xs italic">
			{i18n.t('sidebar.character_stats.dual_wield_penalty')}
		</td>
	</tr>
);
