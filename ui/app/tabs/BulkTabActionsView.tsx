import i18n from '@i18n/config';
import { REPO_RELEASES_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';

export interface BulkTabActionsViewProps {
	showNativeDownload: boolean;
	onImportBags: () => void;
	onImportFavorites: () => void;
	onClear: () => void;
}

export const BulkTabActionsView = ({ showNativeDownload, onImportBags, onImportFavorites, onClear }: BulkTabActionsViewProps) => (
	<>
		<div>
			{showNativeDownload && (
				<p className="mb-0">
					<Button href={REPO_RELEASES_URL} target="_blank" variant="unstyled">
						<Icon name="gauge-high" className="mr-1" />
						{i18n.t('bulk_tab.download_native')}
					</Button>
				</p>
			)}
		</div>
		<div className="flex grid-flow-col gap-2" data-testid="bulk-gear-actions">
			<Button variant="secondary" onClick={onImportBags}>
				<Icon name="download" style="base" /> {i18n.t('bulk_tab.actions.import_bags')}
			</Button>
			<Button variant="secondary" onClick={onImportFavorites}>
				<Icon name="download" style="base" /> {i18n.t('bulk_tab.actions.import_favorites')}
			</Button>
			<Button variant="danger" className="ml-auto" onClick={onClear}>
				<Icon name="xmark" />
				{i18n.t('bulk_tab.actions.clear_items')}
			</Button>
		</div>
	</>
);
