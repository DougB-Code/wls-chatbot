---
name: ui-style-consistency
description: Maintain consistent UI styling for Lit web components and plain CSS in the Glint desktop frontend. Use this skill for any request that touches CSS design or visual layout, including adding/moving form elements, adding buttons, resizing typography, adjusting spacing, or aligning styles across components. Always reuse existing tokens, styles, and components instead of creating new styling systems.
---

# Ui Style Consistency

## Overview

Keep UI styling consistent with existing Glint desktop frontend conventions by reusing current tokens, styles, and components. The goal is to make changes that blend into the existing design language without introducing new styling systems.

## Core Workflow

1. Identify the existing component and its feature folder; do not create duplicate/competing implementations.
2. Locate the closest matching UI element in the same color zone (same panel/surface/section) and reuse its classes, tokens, and structure.
3. Update the existing `*-styles.ts` or `*.css` file in that feature folder; do not introduce new top-level style systems.
4. Prefer shared UI components in `desktop/frontend/src/shared/components` before inventing one-offs.
5. If a form element is added, match the styling of other form elements in that color zone (same tokens, borders, radii, spacing, and typography).
6. Use BEM naming (`gl-block__element--modifier`) and the `gl-` prefix for blocks.

## Lit + CSS Conventions

- Keep styling in the existing `*-styles.ts` files for Lit components and import them in the component.
- Avoid inline styles; use class selectors and tokens.
- Use `:host { display: block; }` (or inline-flex) instead of `display: contents` for interactive elements.
- Use existing CSS custom properties from `desktop/frontend/src/tokens.css` and `desktop/frontend/src/layout/layout-tokens.css` before introducing new ones.

## Consistency Checklist

- Reuse existing tokens and component styles; avoid new palettes or fonts.
- Match spacing, radii, and typography scale in the same area of the UI.
- Keep changes scoped to the requested files only.
- Report any existing code that deviates from project styling guidelines.

## References

- Style and token file locations: `references/style-paths.md`
