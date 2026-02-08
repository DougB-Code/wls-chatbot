export namespace catalog {
	
	export class RoleAssignmentSummary {
	    roleId: string;
	    modelCatalogEntryId: string;
	    modelLabel: string;
	    assignedBy: string;
	    createdAt: number;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RoleAssignmentSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.roleId = source["roleId"];
	        this.modelCatalogEntryId = source["modelCatalogEntryId"];
	        this.modelLabel = source["modelLabel"];
	        this.assignedBy = source["assignedBy"];
	        this.createdAt = source["createdAt"];
	        this.enabled = source["enabled"];
	    }
	}
	export class RoleConstraints {
	    maxCostTier?: string;
	    maxLatencyTier?: string;
	    minReliabilityTier?: string;
	
	    static createFrom(source: any = {}) {
	        return new RoleConstraints(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxCostTier = source["maxCostTier"];
	        this.maxLatencyTier = source["maxLatencyTier"];
	        this.minReliabilityTier = source["minReliabilityTier"];
	    }
	}
	export class RoleRequirements {
	    requiredInputModalities: string[];
	    requiredOutputModalities: string[];
	    requiresStreaming: boolean;
	    requiresToolCalling: boolean;
	    requiresStructuredOutput: boolean;
	    requiresVision: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RoleRequirements(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.requiredInputModalities = source["requiredInputModalities"];
	        this.requiredOutputModalities = source["requiredOutputModalities"];
	        this.requiresStreaming = source["requiresStreaming"];
	        this.requiresToolCalling = source["requiresToolCalling"];
	        this.requiresStructuredOutput = source["requiresStructuredOutput"];
	        this.requiresVision = source["requiresVision"];
	    }
	}
	export class RoleSummary {
	    id: string;
	    name: string;
	    requirements: RoleRequirements;
	    constraints: RoleConstraints;
	    assignments: RoleAssignmentSummary[];
	
	    static createFrom(source: any = {}) {
	        return new RoleSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.requirements = this.convertValues(source["requirements"], RoleRequirements);
	        this.constraints = this.convertValues(source["constraints"], RoleConstraints);
	        this.assignments = this.convertValues(source["assignments"], RoleAssignmentSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ModelSummary {
	    id: string;
	    endpointId: string;
	    modelId: string;
	    displayName: string;
	    availabilityState: string;
	    contextWindow: number;
	    costTier: string;
	    supportsStreaming: boolean;
	    supportsToolCalling: boolean;
	    supportsStructuredOutput: boolean;
	    supportsVision: boolean;
	    inputModalities: string[];
	    outputModalities: string[];
	
	    static createFrom(source: any = {}) {
	        return new ModelSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.endpointId = source["endpointId"];
	        this.modelId = source["modelId"];
	        this.displayName = source["displayName"];
	        this.availabilityState = source["availabilityState"];
	        this.contextWindow = source["contextWindow"];
	        this.costTier = source["costTier"];
	        this.supportsStreaming = source["supportsStreaming"];
	        this.supportsToolCalling = source["supportsToolCalling"];
	        this.supportsStructuredOutput = source["supportsStructuredOutput"];
	        this.supportsVision = source["supportsVision"];
	        this.inputModalities = source["inputModalities"];
	        this.outputModalities = source["outputModalities"];
	    }
	}
	export class EndpointSummary {
	    id: string;
	    providerId: string;
	    providerName: string;
	    providerDisplayName: string;
	    displayName: string;
	    adapterType: string;
	    baseUrl: string;
	    routeKind: string;
	    originProvider: string;
	    originRouteLabel: string;
	    lastTestAt: number;
	    lastTestOk: boolean;
	    lastError?: string;
	    models: ModelSummary[];
	
	    static createFrom(source: any = {}) {
	        return new EndpointSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.providerId = source["providerId"];
	        this.providerName = source["providerName"];
	        this.providerDisplayName = source["providerDisplayName"];
	        this.displayName = source["displayName"];
	        this.adapterType = source["adapterType"];
	        this.baseUrl = source["baseUrl"];
	        this.routeKind = source["routeKind"];
	        this.originProvider = source["originProvider"];
	        this.originRouteLabel = source["originRouteLabel"];
	        this.lastTestAt = source["lastTestAt"];
	        this.lastTestOk = source["lastTestOk"];
	        this.lastError = source["lastError"];
	        this.models = this.convertValues(source["models"], ModelSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProviderSummary {
	    id: string;
	    name: string;
	    displayName: string;
	    adapterType: string;
	    trustMode: string;
	    baseUrl: string;
	    lastTestAt: number;
	    lastTestOk: boolean;
	    lastError?: string;
	    lastDiscoveryAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ProviderSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.adapterType = source["adapterType"];
	        this.trustMode = source["trustMode"];
	        this.baseUrl = source["baseUrl"];
	        this.lastTestAt = source["lastTestAt"];
	        this.lastTestOk = source["lastTestOk"];
	        this.lastError = source["lastError"];
	        this.lastDiscoveryAt = source["lastDiscoveryAt"];
	    }
	}
	export class CatalogOverview {
	    providers: ProviderSummary[];
	    endpoints: EndpointSummary[];
	    roles: RoleSummary[];
	
	    static createFrom(source: any = {}) {
	        return new CatalogOverview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providers = this.convertValues(source["providers"], ProviderSummary);
	        this.endpoints = this.convertValues(source["endpoints"], EndpointSummary);
	        this.roles = this.convertValues(source["roles"], RoleSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class RoleAssignmentResult {
	    missingModalities?: string[];
	    missingFeatures?: string[];
	
	    static createFrom(source: any = {}) {
	        return new RoleAssignmentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.missingModalities = source["missingModalities"];
	        this.missingFeatures = source["missingFeatures"];
	    }
	}
	
	
	

}

export namespace core {
	
	export class CredentialField {
	    name: string;
	    label: string;
	    required: boolean;
	    secret: boolean;
	    placeholder?: string;
	    help?: string;
	
	    static createFrom(source: any = {}) {
	        return new CredentialField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.label = source["label"];
	        this.required = source["required"];
	        this.secret = source["secret"];
	        this.placeholder = source["placeholder"];
	        this.help = source["help"];
	    }
	}
	export class Model {
	    id: string;
	    name: string;
	    contextWindow: number;
	    supportsStreaming: boolean;
	    supportsTools: boolean;
	    supportsVision: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Model(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.contextWindow = source["contextWindow"];
	        this.supportsStreaming = source["supportsStreaming"];
	        this.supportsTools = source["supportsTools"];
	        this.supportsVision = source["supportsVision"];
	    }
	}

}

export namespace domain {
	
	export class ActionExecution {
	    id: string;
	    toolName: string;
	    description: string;
	    args: Record<string, any>;
	    status: string;
	    result?: string;
	    startedAt?: number;
	    completedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new ActionExecution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.toolName = source["toolName"];
	        this.description = source["description"];
	        this.args = source["args"];
	        this.status = source["status"];
	        this.result = source["result"];
	        this.startedAt = source["startedAt"];
	        this.completedAt = source["completedAt"];
	    }
	}
	export class Artifact {
	    id: string;
	    name: string;
	    type: string;
	    content: string;
	    language?: string;
	    version: number;
	    createdAt: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Artifact(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.language = source["language"];
	        this.version = source["version"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class Block {
	    type: string;
	    content: string;
	    language?: string;
	    artifact?: Artifact;
	    action?: ActionExecution;
	    isCollapsed?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Block(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.content = source["content"];
	        this.language = source["language"];
	        this.artifact = this.convertValues(source["artifact"], Artifact);
	        this.action = this.convertValues(source["action"], ActionExecution);
	        this.isCollapsed = source["isCollapsed"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConversationSettings {
	    provider: string;
	    model: string;
	    temperature?: number;
	    maxTokens?: number;
	    systemPrompt?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversationSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.temperature = source["temperature"];
	        this.maxTokens = source["maxTokens"];
	        this.systemPrompt = source["systemPrompt"];
	    }
	}
	export class MessageMetadata {
	    provider?: string;
	    model?: string;
	    tokensIn?: number;
	    tokensOut?: number;
	    tokensTotal?: number;
	    latencyMs?: number;
	    finishReason?: string;
	    statusCode?: number;
	    errorMessage?: string;
	
	    static createFrom(source: any = {}) {
	        return new MessageMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.tokensIn = source["tokensIn"];
	        this.tokensOut = source["tokensOut"];
	        this.tokensTotal = source["tokensTotal"];
	        this.latencyMs = source["latencyMs"];
	        this.finishReason = source["finishReason"];
	        this.statusCode = source["statusCode"];
	        this.errorMessage = source["errorMessage"];
	    }
	}
	export class Message {
	    id: string;
	    conversationId: string;
	    role: string;
	    blocks: Block[];
	    timestamp: number;
	    isStreaming?: boolean;
	    metadata?: MessageMetadata;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.conversationId = source["conversationId"];
	        this.role = source["role"];
	        this.blocks = this.convertValues(source["blocks"], Block);
	        this.timestamp = source["timestamp"];
	        this.isStreaming = source["isStreaming"];
	        this.metadata = this.convertValues(source["metadata"], MessageMetadata);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Conversation {
	    id: string;
	    title: string;
	    messages: Message[];
	    settings: ConversationSettings;
	    createdAt: number;
	    updatedAt: number;
	    isArchived: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.messages = this.convertValues(source["messages"], Message);
	        this.settings = this.convertValues(source["settings"], ConversationSettings);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.isArchived = source["isArchived"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ConversationSummary {
	    id: string;
	    title: string;
	    lastMessage?: string;
	    messageCount: number;
	    updatedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new ConversationSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.lastMessage = source["lastMessage"];
	        this.messageCount = source["messageCount"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	
	
	export class Notification {
	    id: number;
	    type: string;
	    title: string;
	    message: string;
	    createdAt: number;
	    readAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new Notification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.message = source["message"];
	        this.createdAt = source["createdAt"];
	        this.readAt = source["readAt"];
	    }
	}

}

export namespace interfaces {
	
	export class EditImageRequest {
	    providerName: string;
	    modelName?: string;
	    prompt: string;
	    imagePath: string;
	    maskPath?: string;
	    n?: number;
	    size?: string;
	
	    static createFrom(source: any = {}) {
	        return new EditImageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerName = source["providerName"];
	        this.modelName = source["modelName"];
	        this.prompt = source["prompt"];
	        this.imagePath = source["imagePath"];
	        this.maskPath = source["maskPath"];
	        this.n = source["n"];
	        this.size = source["size"];
	    }
	}
	export class GenerateImageRequest {
	    providerName: string;
	    modelName?: string;
	    prompt: string;
	    n?: number;
	    size?: string;
	    quality?: string;
	    style?: string;
	    responseFormat?: string;
	    user?: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateImageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerName = source["providerName"];
	        this.modelName = source["modelName"];
	        this.prompt = source["prompt"];
	        this.n = source["n"];
	        this.size = source["size"];
	        this.quality = source["quality"];
	        this.style = source["style"];
	        this.responseFormat = source["responseFormat"];
	        this.user = source["user"];
	    }
	}
	export class ImageBinaryResult {
	    bytes: number[];
	    revisedPrompt?: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageBinaryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bytes = source["bytes"];
	        this.revisedPrompt = source["revisedPrompt"];
	    }
	}
	export class ModelCapabilities {
	    supportsStreaming: boolean;
	    supportsToolCalling: boolean;
	    supportsStructuredOutput: boolean;
	    supportsVision: boolean;
	    inputModalities: string[];
	    outputModalities: string[];
	    capabilityIds: string[];
	    systemTags?: string[];
	
	    static createFrom(source: any = {}) {
	        return new ModelCapabilities(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supportsStreaming = source["supportsStreaming"];
	        this.supportsToolCalling = source["supportsToolCalling"];
	        this.supportsStructuredOutput = source["supportsStructuredOutput"];
	        this.supportsVision = source["supportsVision"];
	        this.inputModalities = source["inputModalities"];
	        this.outputModalities = source["outputModalities"];
	        this.capabilityIds = source["capabilityIds"];
	        this.systemTags = source["systemTags"];
	    }
	}
	export class ModelListFilter {
	    source?: string;
	    requiredInputModalities?: string[];
	    requiredOutputModalities?: string[];
	    requiredCapabilityIds?: string[];
	    requiredSystemTags?: string[];
	    requiresStreaming?: boolean;
	    requiresToolCalling?: boolean;
	    requiresStructuredOutput?: boolean;
	    requiresVision?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelListFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.source = source["source"];
	        this.requiredInputModalities = source["requiredInputModalities"];
	        this.requiredOutputModalities = source["requiredOutputModalities"];
	        this.requiredCapabilityIds = source["requiredCapabilityIds"];
	        this.requiredSystemTags = source["requiredSystemTags"];
	        this.requiresStreaming = source["requiresStreaming"];
	        this.requiresToolCalling = source["requiresToolCalling"];
	        this.requiresStructuredOutput = source["requiresStructuredOutput"];
	        this.requiresVision = source["requiresVision"];
	    }
	}
	export class ModelSummary {
	    id: string;
	    modelId: string;
	    displayName: string;
	    providerName: string;
	    source: string;
	    approved: boolean;
	    availabilityState: string;
	    contextWindow: number;
	    costTier: string;
	    capabilities: ModelCapabilities;
	
	    static createFrom(source: any = {}) {
	        return new ModelSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.modelId = source["modelId"];
	        this.displayName = source["displayName"];
	        this.providerName = source["providerName"];
	        this.source = source["source"];
	        this.approved = source["approved"];
	        this.availabilityState = source["availabilityState"];
	        this.contextWindow = source["contextWindow"];
	        this.costTier = source["costTier"];
	        this.capabilities = this.convertValues(source["capabilities"], ModelCapabilities);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SyncModelsResult {
	    path: string;
	    imported: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SyncModelsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.imported = source["imported"];
	    }
	}

}

export namespace logger {
	
	export class LogEntry {
	    level: string;
	    message: string;
	    fields: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.fields = source["fields"];
	    }
	}

}

export namespace notifications {
	
	export class NotificationPayload {
	    type: string;
	    title: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new NotificationPayload(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.title = source["title"];
	        this.message = source["message"];
	    }
	}

}

export namespace provider {
	
	export class Status {
	    ok: boolean;
	    message?: string;
	    checkedAt: number;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.checkedAt = source["checkedAt"];
	    }
	}
	export class Info {
	    name: string;
	    displayName: string;
	    credentialFields?: core.CredentialField[];
	    credentialValues?: Record<string, string>;
	    models: core.Model[];
	    resources: core.Model[];
	    isConnected: boolean;
	    isActive: boolean;
	    status?: Status;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.credentialFields = this.convertValues(source["credentialFields"], core.CredentialField);
	        this.credentialValues = source["credentialValues"];
	        this.models = this.convertValues(source["models"], core.Model);
	        this.resources = this.convertValues(source["resources"], core.Model);
	        this.isConnected = source["isConnected"];
	        this.isActive = source["isActive"];
	        this.status = this.convertValues(source["status"], Status);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

