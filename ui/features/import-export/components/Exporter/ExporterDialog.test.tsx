import i18n from '@i18n/config';
import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// The exporters reach the addon importer, which asks GitHub for its version when its module loads.
// Answered offline here, in `vi.hoisted` so it runs before the imports, no request is left open
// when the window closes.
vi.hoisted(() => {
	(globalThis as unknown as { fetch: unknown }).fetch = () => Promise.reject(new Error('offline'));
});

import { renderWithHost } from '../../../../testing/react';
import { EXPORT_CATEGORY_OPTIONS } from '../../exporters';
import { exporterDialog } from './ExporterDialog';

const DOWNLOAD = i18n.t('export.json.download_button');

const getData = () => '{"player":{}}';
const TestExporterDialog = exporterDialog({ title: 'Test Export', getData });
const FullExporterDialog = exporterDialog({ title: 'Full Export', allowDownload: true, selectCategories: true, getData });

describe('exporterDialog', () => {
	it('names the component after the definition', () => {
		expect(TestExporterDialog.displayName).toBe('ExporterDialog(Test Export)');
	});

	it('opens the exporter under the definition title with its data', () => {
		renderWithHost(<TestExporterDialog open onOpenChange={() => {}} />);
		expect(screen.getByTestId('sim-dialog-title').textContent).toBe('Test Export');
		expect(screen.getByTestId<HTMLTextAreaElement>('exporter-textarea').value).toBe('{"player":{}}');
	});

	it('leaves the category pickers and the download button out unless the definition asks', () => {
		renderWithHost(<TestExporterDialog open onOpenChange={() => {}} />);
		expect(screen.queryAllByRole('checkbox')).toHaveLength(0);
		expect(screen.queryByRole('button', { name: DOWNLOAD })).toBeNull();
	});

	it('shows the category pickers and the download button when the definition asks', () => {
		renderWithHost(<FullExporterDialog open onOpenChange={() => {}} />);
		expect(screen.getAllByRole('checkbox')).toHaveLength(EXPORT_CATEGORY_OPTIONS.length);
		expect(screen.getByRole('button', { name: DOWNLOAD })).toBeTruthy();
	});
});
