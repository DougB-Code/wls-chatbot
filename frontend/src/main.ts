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
    const startup = document.createElement('div');
    startup.className = 'wls-startup';
    startup.textContent = 'Loading WLS ChatBot...';

    const appShell = document.createElement('wls-app-shell');
    appShell.style.visibility = 'hidden';
    root.append(startup, appShell);

    void initAppController(root).finally(() => {
        startup.remove();
        appShell.style.visibility = 'visible';
        document.body.dataset.appReady = 'true';
    });
}
