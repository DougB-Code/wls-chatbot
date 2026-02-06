/**
 * appController.ts wires UI events to policy actions and initializes app policies.
 * frontend/src/app/application/appController.ts
 */

import { activeConversation, approveAction, rejectAction } from '../../features/chat/state/chatSignals';
import { setPreferredModelId } from '../../features/chat/state/chatPreferences';
import { effectiveModelId } from '../../features/chat/application/chatSelectors';
import { createNewConversation, deleteConversation, initChatPolicy, purgeConversation, restoreConversation, selectConversation, sendMessage, setConversationModel, setConversationProvider, stopStream } from '../../features/chat/application/chatPolicy';
import { initProviderEvents, refreshProviders, connectProvider, configureProvider, disconnectProvider, refreshProviderResources, setActiveProvider } from '../../features/settings/application/providerPolicy';
import { initCatalogEvents, refreshCatalogOverview, refreshCatalogEndpoint, testCatalogEndpoint, saveCatalogRole, deleteCatalogRole, assignCatalogRole, unassignCatalogRole } from '../../features/settings/application/catalogPolicy';
import { catalogOverview } from '../../features/settings/state/catalogSignals';
import { initToastEvents } from './toastPolicy';
import { notifyError } from './notificationPolicy';
import { refreshNotifications } from '../../features/notifications/application/notificationPolicy';
import type { RoleSummary } from '../../types/catalog';

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
    initCatalogEvents();
    void refreshProviders().catch((err) => {
        notifyError((err as Error).message || 'Failed to load providers', 'Startup failed');
    });
    void refreshCatalogOverview().catch((err) => {
        notifyError((err as Error).message || 'Failed to load catalog', 'Startup failed');
    });
    void initChatPolicy().catch((err) => {
        notifyError((err as Error).message || 'Failed to initialize chat', 'Startup failed');
    });
    void refreshNotifications().catch((err) => {
        notifyError((err as Error).message || 'Failed to load notifications', 'Startup failed');
    });
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
            if (!conversation) {
                return;
            }

            if (conversation.settings?.provider !== detail.provider) {
                await setConversationProvider(conversation.id, detail.provider);
            }

            if (conversation.settings?.model !== selectedModel) {
                await setConversationModel(conversation.id, selectedModel);
            }
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
        const detail = (event as CustomEvent<{ name: string; credentials: Record<string, string> }>).detail;
        if (!detail?.name || !detail?.credentials || Object.keys(detail.credentials).length === 0) return;
        void connectProvider(detail.name, detail.credentials).catch((err) => {
            notifyError((err as Error).message || `Failed to connect ${detail.name}`, 'Provider connect failed');
        });
    });

    root.addEventListener('provider-configure', (event: Event) => {
        const detail = (event as CustomEvent<{ name: string; credentials: Record<string, string> }>).detail;
        if (!detail?.name || !detail?.credentials || Object.keys(detail.credentials).length === 0) return;
        void configureProvider(detail.name, detail.credentials).catch((err) => {
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
        void (async () => {
            await refreshProviderResources(detail.name);
            const endpointIds = (catalogOverview.value?.endpoints ?? [])
                .filter((endpoint) => endpoint.providerName === detail.name)
                .map((endpoint) => endpoint.id);
            if (endpointIds.length === 0) {
                await refreshCatalogOverview();
                return;
            }
            for (const endpointId of endpointIds) {
                await refreshCatalogEndpoint(endpointId);
            }
        })().catch((err) => {
            notifyError((err as Error).message || `Failed to refresh ${detail.name}`, 'Provider refresh failed');
        });
    });

    root.addEventListener('catalog-endpoint-test', (event: Event) => {
        const detail = (event as CustomEvent<{ endpointId: string }>).detail;
        if (!detail?.endpointId) return;
        void testCatalogEndpoint(detail.endpointId).catch((err) => {
            notifyError((err as Error).message || 'Failed to test endpoint', 'Endpoint test failed');
        });
    });

    root.addEventListener('catalog-endpoint-refresh', (event: Event) => {
        const detail = (event as CustomEvent<{ endpointId: string }>).detail;
        if (!detail?.endpointId) return;
        void refreshCatalogEndpoint(detail.endpointId).catch((err) => {
            notifyError((err as Error).message || 'Failed to refresh endpoint', 'Endpoint refresh failed');
        });
    });

    root.addEventListener('catalog-role-save', (event: Event) => {
        const detail = (event as CustomEvent<{ role: RoleSummary }>).detail;
        if (!detail?.role) return;
        void saveCatalogRole(detail.role).catch((err) => {
            notifyError((err as Error).message || 'Failed to save role', 'Role save failed');
        });
    });

    root.addEventListener('catalog-role-delete', (event: Event) => {
        const detail = (event as CustomEvent<{ roleId: string }>).detail;
        if (!detail?.roleId) return;
        void deleteCatalogRole(detail.roleId).catch((err) => {
            notifyError((err as Error).message || 'Failed to delete role', 'Role delete failed');
        });
    });

    root.addEventListener('catalog-role-assign', (event: Event) => {
        const detail = (event as CustomEvent<{ roleId: string; modelEntryId: string }>).detail;
        if (!detail?.roleId || !detail?.modelEntryId) return;
        void assignCatalogRole(detail.roleId, detail.modelEntryId, 'user').catch((err) => {
            notifyError((err as Error).message || 'Failed to assign role', 'Role assignment failed');
        });
    });

    root.addEventListener('catalog-role-unassign', (event: Event) => {
        const detail = (event as CustomEvent<{ roleId: string; modelEntryId: string }>).detail;
        if (!detail?.roleId || !detail?.modelEntryId) return;
        void unassignCatalogRole(detail.roleId, detail.modelEntryId).catch((err) => {
            notifyError((err as Error).message || 'Failed to remove assignment', 'Role assignment failed');
        });
    });
}
