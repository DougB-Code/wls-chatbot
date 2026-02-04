// eslint.config.js
// @ts-check

import eslint from '@eslint/js';
import { defineConfig } from 'eslint/config';
import tseslint from 'typescript-eslint';
import prettier from 'eslint-config-prettier';

export default defineConfig(
  // Global ignores (flat-config replacement for .eslintignore)
  {
    ignores: [
      '**/node_modules/**',
      '**/dist/**',
      '**/build/**',
      '**/coverage/**',
      '**/.turbo/**',
      '**/.next/**',
      '**/.output/**',
      '**/out/**',
    ],
  },

  // Base recommended rules
  eslint.configs.recommended,

  // TypeScript rules (recommended + stricter bug-catching + stylistic consistency)
  tseslint.configs.recommended,
  tseslint.configs.strict,
  tseslint.configs.stylistic,

  // Naming conventions (the “enforced” version of the conventions we discussed)
  {
    files: ['**/*.{ts,tsx,js,jsx,mjs,cjs}'],
    rules: {
      // Enforce identifier casing across the codebase.
      // Rule docs: https://typescript-eslint.io/rules/naming-convention
      '@typescript-eslint/naming-convention': [
        'error',

        // Default: camelCase (variables, functions, params, properties, etc.)
        {
          selector: 'default',
          format: ['camelCase'],
          leadingUnderscore: 'allow',
          trailingUnderscore: 'forbid',
        },

        // Types: PascalCase (classes, interfaces, type aliases, enums, etc.)
        {
          selector: 'typeLike',
          format: ['PascalCase'],
        },

        // Enum members: PascalCase or UPPER_CASE (pick either style per enum)
        {
          selector: 'enumMember',
          format: ['PascalCase', 'UPPER_CASE'],
        },

        // Consts: allow camelCase OR UPPER_CASE
        {
          selector: 'variable',
          modifiers: ['const'],
          format: ['camelCase', 'UPPER_CASE'],
          leadingUnderscore: 'allow',
        },

        // Parameters: camelCase; allow leading underscore for intentionally-unused args
        {
          selector: 'parameter',
          format: ['camelCase'],
          leadingUnderscore: 'allow',
        },

        // Properties: camelCase; allow leading underscore for private-ish fields if you prefer
        {
          selector: 'property',
          format: ['camelCase'],
          leadingUnderscore: 'allow',
        },

        // Methods / functions: camelCase
        {
          selector: 'function',
          format: ['camelCase'],
        },
      ],
    },
  },

  // Disable ESLint formatting rules that conflict with Prettier.
  // (Run Prettier separately; ESLint focuses on correctness/quality.)
  prettier,
);
