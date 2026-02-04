/**
 * appController.ts wires UI events to policy actions and initializes app policies.
 * frontend/src/policy/appController.ts
 */

import { activeConversation, approveAction, rejectAction } from '../store/signals';
import { setPreferredModelId } from '../store/chatPreferences';
import { effectiveModelId } from './chatSelectors';
import { createNewConversation, deleteConversation, initChatPolicy, purgeConversation, restoreConversation, selectConversation, sendMessage, setConversationModel, stopStream } from './chatPolicy';
import { initProviderEvents, refreshProviders, connectProvider, configureProvider, disconnectProvider, refreshProviderResources, setActiveProvider } from './providerPolicy';
import { initToastEvents } from './toastPolicy';
import { notifyError } from './notificationPolicy';

/**
 * initialize app policies and wire UI event handlers.
 */
export async function initAppController(root: HTMLElement): Promise<void> {
    attachUiHandlers(root);
    await bootstrapPolicies();
}

/**
 * initialize policy subsystems and hydrate initial state.
 */
async function bootstrapPolicies(): Promise<void> {
    initToastEvents();
    initProviderEvents();
    const results = await Promise.allSettled([refreshProviders(), initChatPolicy()]);

    for (const result of results) {
        if (result.status === 'rejected') {
            notifyError((result.reason as Error)?.message || 'Failed to initialize the app', 'Startup failed');
        }
    }
}

/**
 * connect UI events to policy actions.
 */
function attachUiHandlers(root: HTMLElement): void {
    root.addEventListener('chat-send', (event: Event) => {
        const detail = (event as CustomEvent<{ content: string; attachments: File[] }>).detail;
        if (!detail?.content && (!detail?.attachments || detail.attachments.length === 0)) {
            return;
        }
        const conversation = activeConversation.value ?? null;
        void sendMessage(detail.content, detail.attachments ?? [], undefined, conversation).catch((err) => {
            notifyError((err as Error).message || 'Failed to send message', 'Send failed');
        });
    });

    root.addEventListener('chat-stop-stream', () => {
        void stopStream().catch((err) => {
            notifyError((err as Error).message || 'Failed to stop stream', 'Stop failed');
        });
    });

    root.addEventListener('chat-create', () => {
        void createNewConversation().catch((err) => {
            notifyError((err as Error).message || 'Failed to create conversation', 'Create failed');
        });
    });

    root.addEventListener('chat-select', (event: Event) => {
        const detail = (event as CustomEvent<{ conversationId: string }>).detail;
        if (!detail?.conversationId) return;
        void selectConversation(detail.conversationId).catch((err) => {
            notifyError((err as Error).message || 'Failed to load conversation', 'Load failed');
        });
    });

    root.addEventListener('chat-delete', (event: Event) => {
        const detail = (event as CustomEvent<{ conversationId: string }>).detail;
        if (!detail?.conversationId) return;
        void deleteConversation(detail.conversationId).catch((err) => {
            notifyError((err as Error).message || 'Failed to delete conversation', 'Delete failed');
        });
    });

    root.addEventListener('chat-restore', (event: Event) => {
        const detail = (event as CustomEvent<{ conversationId: string }>).detail;
        if (!detail?.conversationId) return;
        void restoreConversation(detail.conversationId).catch((err) => {
            notifyError((err as Error).message || 'Failed to restore conversation', 'Restore failed');
        });
    });

    root.addEventListener('chat-purge', (event: Event) => {
        const detail = (event as CustomEvent<{ conversationId: string }>).detail;
        if (!detail?.conversationId) return;
        void purgeConversation(detail.conversationId).catch((err) => {
            notifyError((err as Error).message || 'Failed to permanently delete conversation', 'Delete failed');
        });
    });

    root.addEventListener('chat-model-select', (event: Event) => {
        const detail = (event as CustomEvent<{ model: string }>).detail;
        if (!detail?.model) return;
        setPreferredModelId(detail.model);
        const conversation = activeConversation.value;
        if (!conversation) return;
        void setConversationModel(conversation.id, detail.model).catch((err) => {
            notifyError((err as Error).message || 'Failed to update model', 'Model update failed');
        });
    });

    root.addEventListener('chat-provider-select', (event: Event) => {
        const detail = (event as CustomEvent<{ provider: string }>).detail;
        if (!detail?.provider) return;

        void setActiveProvider(detail.provider).then(async () => {
            const selectedModel = effectiveModelId.value;
            if (!selectedModel) {
                return;
            }

            setPreferredModelId(selectedModel);
            const conversation = activeConversation.value;
            if (!conversation || conversation.settings?.model === selectedModel) {
                return;
            }

            await setConversationModel(conversation.id, selectedModel);
        }).catch((err) => {
            notifyError((err as Error).message || `Failed to set active provider: ${detail.provider}`, 'Provider update failed');
        });
    });

    root.addEventListener('chat-action-approve', (event: Event) => {
        const detail = (event as CustomEvent<{ actionId: string }>).detail;
        const conversation = activeConversation.value;
        if (!conversation || !detail?.actionId) return;
        approveAction(conversation.id, detail.actionId);
    });

    root.addEventListener('chat-action-reject', (event: Event) => {
        const detail = (event as CustomEvent<{ actionId: string }>).detail;
        const conversation = activeConversation.value;
        if (!conversation || !detail?.actionId) return;
        rejectAction(conversation.id, detail.actionId);
    });

    root.addEventListener('provider-connect', (event: Event) => {
        const detail = (event as CustomEvent<{ name: string; apiKey: string }>).detail;
        if (!detail?.name || !detail?.apiKey) return;
        void connectProvider(detail.name, detail.apiKey).catch((err) => {
            notifyError((err as Error).message || `Failed to connect ${detail.name}`, 'Provider connect failed');
        });
    });

    root.addEventListener('provider-configure', (event: Event) => {
        const detail = (event as CustomEvent<{ name: string; apiKey: string }>).detail;
        if (!detail?.name || !detail?.apiKey) return;
        void configureProvider(detail.name, detail.apiKey).catch((err) => {
            notifyError((err as Error).message || `Failed to update ${detail.name}`, 'Provider update failed');
        });
    });

    root.addEventListener('provider-disconnect', (event: Event) => {
        const detail = (event as CustomEvent<{ name: string }>).detail;
        if (!detail?.name) return;
        void disconnectProvider(detail.name).catch((err) => {
            notifyError((err as Error).message || `Failed to disconnect ${detail.name}`, 'Provider disconnect failed');
        });
    });

    root.addEventListener('provider-refresh', (event: Event) => {
        const detail = (event as CustomEvent<{ name: string }>).detail;
        if (!detail?.name) return;
        void refreshProviderResources(detail.name).catch((err) => {
            notifyError((err as Error).message || `Failed to refresh ${detail.name}`, 'Provider refresh failed');
        });
    });
}
