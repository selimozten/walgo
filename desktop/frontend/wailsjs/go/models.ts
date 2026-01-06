export namespace api {
	
	export class AIConfigResult {
	    success: boolean;
	    enabled: boolean;
	    currentProvider: string;
	    currentModel: string;
	    configuredProviders: string[];
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfigResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.enabled = source["enabled"];
	        this.currentProvider = source["currentProvider"];
	        this.currentModel = source["currentModel"];
	        this.configuredProviders = source["configuredProviders"];
	        this.error = source["error"];
	    }
	}
	export class AIConfigureParams {
	    provider: string;
	    apiKey: string;
	    baseURL: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfigureParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.apiKey = source["apiKey"];
	        this.baseURL = source["baseURL"];
	        this.model = source["model"];
	    }
	}
	export class AICreateSiteParams {
	    parentDir: string;
	    siteName: string;
	    siteType: string;
	    description: string;
	    audience: string;
	    features: string;
	    themeName: string;
	    themeUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new AICreateSiteParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentDir = source["parentDir"];
	        this.siteName = source["siteName"];
	        this.siteType = source["siteType"];
	        this.description = source["description"];
	        this.audience = source["audience"];
	        this.features = source["features"];
	        this.themeName = source["themeName"];
	        this.themeUrl = source["themeUrl"];
	    }
	}
	export class AICreateSiteResult {
	    success: boolean;
	    sitePath: string;
	    filesCreated: number;
	    error: string;
	    steps: string[];
	
	    static createFrom(source: any = {}) {
	        return new AICreateSiteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.sitePath = source["sitePath"];
	        this.filesCreated = source["filesCreated"];
	        this.error = source["error"];
	        this.steps = source["steps"];
	    }
	}
	export class ArchiveProjectResult {
	    success: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ArchiveProjectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class DoctorCheck {
	    name: string;
	    status: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DoctorCheck(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	    }
	}
	export class DoctorSummary {
	    issues: number;
	    warnings: number;
	
	    static createFrom(source: any = {}) {
	        return new DoctorSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.issues = source["issues"];
	        this.warnings = source["warnings"];
	    }
	}
	export class DoctorResult {
	    success: boolean;
	    checks: DoctorCheck[];
	    summary: DoctorSummary;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new DoctorResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.checks = this.convertValues(source["checks"], DoctorCheck);
	        this.summary = this.convertValues(source["summary"], DoctorSummary);
	        this.error = source["error"];
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
	
	export class EditProjectParams {
	    projectId: number;
	    name: string;
	    category: string;
	    description: string;
	    imageUrl: string;
	    suins: string;
	
	    static createFrom(source: any = {}) {
	        return new EditProjectParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.imageUrl = source["imageUrl"];
	        this.suins = source["suins"];
	    }
	}
	export class EditProjectResult {
	    success: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new EditProjectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class LaunchStep {
	    name: string;
	    status: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LaunchStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	    }
	}
	export class LaunchWizardParams {
	    sitePath: string;
	    network: string;
	    projectName: string;
	    category: string;
	    description: string;
	    epochs: number;
	
	    static createFrom(source: any = {}) {
	        return new LaunchWizardParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.network = source["network"];
	        this.projectName = source["projectName"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.epochs = source["epochs"];
	    }
	}
	export class LaunchWizardResult {
	    success: boolean;
	    objectId: string;
	    steps: LaunchStep[];
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new LaunchWizardResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.objectId = source["objectId"];
	        this.steps = this.convertValues(source["steps"], LaunchStep);
	        this.error = source["error"];
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
	export class NewContentParams {
	    sitePath: string;
	    slug: string;
	    contentType: string;
	
	    static createFrom(source: any = {}) {
	        return new NewContentParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.slug = source["slug"];
	        this.contentType = source["contentType"];
	    }
	}
	export class NewContentResult {
	    success: boolean;
	    filePath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new NewContentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.filePath = source["filePath"];
	        this.error = source["error"];
	    }
	}
	export class Project {
	    id: number;
	    name: string;
	    category: string;
	    description: string;
	    network: string;
	    objectId: string;
	    suinsDomain: string;
	    status: string;
	    deploymentCount: number;
	    // Go type: time
	    lastDeployed: any;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.category = source["category"];
	        this.description = source["description"];
	        this.network = source["network"];
	        this.objectId = source["objectId"];
	        this.suinsDomain = source["suinsDomain"];
	        this.status = source["status"];
	        this.deploymentCount = source["deploymentCount"];
	        this.lastDeployed = this.convertValues(source["lastDeployed"], null);
	        this.createdAt = this.convertValues(source["createdAt"], null);
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
	export class QuickStartParams {
	    parentDir: string;
	    name: string;
	    skipBuild: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QuickStartParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentDir = source["parentDir"];
	        this.name = source["name"];
	        this.skipBuild = source["skipBuild"];
	    }
	}
	export class QuickStartResult {
	    success: boolean;
	    sitePath: string;
	    error: string;
	    hasTheme: boolean;
	    built: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QuickStartResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.sitePath = source["sitePath"];
	        this.error = source["error"];
	        this.hasTheme = source["hasTheme"];
	        this.built = source["built"];
	    }
	}
	export class ServeParams {
	    sitePath: string;
	    port: number;
	    drafts: boolean;
	    expired: boolean;
	    future: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServeParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.port = source["port"];
	        this.drafts = source["drafts"];
	        this.expired = source["expired"];
	        this.future = source["future"];
	    }
	}
	export class ServeResult {
	    success: boolean;
	    url: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ServeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.url = source["url"];
	        this.error = source["error"];
	    }
	}
	export class SetupDepsResult {
	    success: boolean;
	    dependencies: Record<string, string>;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SetupDepsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.dependencies = source["dependencies"];
	        this.error = source["error"];
	    }
	}
	export class StatusResult {
	    success: boolean;
	    objectId: string;
	    resourceCount: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new StatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.objectId = source["objectId"];
	        this.resourceCount = source["resourceCount"];
	        this.error = source["error"];
	    }
	}
	export class UpdateDeploymentParams {
	    sitePath: string;
	    objectId: string;
	    epochs: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateDeploymentParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.objectId = source["objectId"];
	        this.epochs = source["epochs"];
	    }
	}
	export class UpdateDeploymentResult {
	    success: boolean;
	    objectId: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateDeploymentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.objectId = source["objectId"];
	        this.error = source["error"];
	    }
	}
	export class VersionResult {
	    version: string;
	    buildDate: string;
	    goVersion: string;
	    platform: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.buildDate = source["buildDate"];
	        this.goVersion = source["goVersion"];
	        this.platform = source["platform"];
	    }
	}

}

export namespace main {
	
	export class DeployResult {
	    success: boolean;
	    objectId: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new DeployResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.objectId = source["objectId"];
	        this.error = source["error"];
	    }
	}
	export class GenerateContentParams {
	    sitePath: string;
	    contentType: string;
	    topic: string;
	    context: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateContentParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.contentType = source["contentType"];
	        this.topic = source["topic"];
	        this.context = source["context"];
	    }
	}
	export class GenerateContentResult {
	    success: boolean;
	    filePath: string;
	    content: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateContentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.filePath = source["filePath"];
	        this.content = source["content"];
	        this.error = source["error"];
	    }
	}
	export class ImportObsidianParams {
	    sitePath: string;
	    vaultPath: string;
	    includeDrafts: boolean;
	    attachmentDir: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportObsidianParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.vaultPath = source["vaultPath"];
	        this.includeDrafts = source["includeDrafts"];
	        this.attachmentDir = source["attachmentDir"];
	    }
	}
	export class ImportObsidianResult {
	    success: boolean;
	    filesImported: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportObsidianResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.filesImported = source["filesImported"];
	        this.error = source["error"];
	    }
	}
	export class UpdateContentParams {
	    filePath: string;
	    instructions: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateContentParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.instructions = source["instructions"];
	    }
	}
	export class UpdateContentResult {
	    success: boolean;
	    updatedContent: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateContentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.updatedContent = source["updatedContent"];
	        this.error = source["error"];
	    }
	}

}

