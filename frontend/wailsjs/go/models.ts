export namespace chat {
	
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

export namespace ports {
	
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
	    models: ports.Model[];
	    resources: ports.Model[];
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
	        this.models = this.convertValues(source["models"], ports.Model);
	        this.resources = this.convertValues(source["resources"], ports.Model);
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

