/**
 * declare Vite-specific type augmentations for the frontend build.
 */
/// <reference types="vite/client" />

/**
 * model raw string imports for assets using the ?raw query.
 */
declare module '*?raw' {
    const content: string;
    export default content;
}
