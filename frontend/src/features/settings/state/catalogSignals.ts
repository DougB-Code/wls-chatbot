/**
 * catalogSignals.ts stores model catalog and role state for settings.
 * frontend/src/features/settings/state/catalogSignals.ts
 */

import { signal } from '@preact/signals-core';
import type { CatalogOverview } from '../../../types/catalog';

export const catalogOverview = signal<CatalogOverview | null>(null);
export const catalogBusy = signal<Record<string, boolean>>({});
export const catalogError = signal<string | null>(null);

/**
 * replace the catalog overview state.
 */
export function setCatalogOverview(overview: CatalogOverview | null): void {
    catalogOverview.value = overview;
}

/**
 * set busy state for catalog actions.
 */
export function setCatalogBusy(key: string, busy: boolean): void {
    catalogBusy.value = { ...catalogBusy.value, [key]: busy };
}

/**
 * set the catalog error state.
 */
export function setCatalogError(error: string | null): void {
    catalogError.value = error;
}
