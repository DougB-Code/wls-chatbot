/**
 * define Vitest configuration for frontend unit tests.
 * frontend/vitest.config.ts
 */

import { defineConfig } from 'vitest/config';

export default defineConfig({
    test: {
        environment: 'node',
        include: ['src/**/*.test.ts'],
    },
});
