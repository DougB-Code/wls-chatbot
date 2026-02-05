/**
 * wails.ts re-exports generated Wails model types for frontend use.
 * frontend/src/types/wails.ts
 */

import type { chat, ports, provider, notifications } from '../../wailsjs/go/models';

export type ActionExecution = chat.ActionExecution;
export type Artifact = chat.Artifact;
export type Block = chat.Block;
export type Conversation = chat.Conversation;
export type ConversationSettings = chat.ConversationSettings;
export type ConversationSummary = chat.ConversationSummary;
export type Message = chat.Message;
export type MessageMetadata = chat.MessageMetadata;

export type ProviderModel = ports.Model;
export type ProviderInfo = provider.Info;
export type ProviderStatus = provider.Status;
export type Notification = notifications.Notification;
export type NotificationPayload = notifications.NotificationPayload;
