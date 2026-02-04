/**
 * bootstrap the frontend app, mount the shell, and initialize controllers.
 */

import './styles/tokens.css';
import './styles/chat.css';
import './style.css';

import './AppShell';
import { initAppController } from './policy/appController';

const root = document.querySelector<HTMLElement>('#app');

if (root) {
    const appShell = document.createElement('wls-app-shell');
    root.appendChild(appShell);
    initAppController(appShell);
}
