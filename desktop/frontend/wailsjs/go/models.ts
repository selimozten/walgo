export namespace ai {
	
	export class ContentFileInfo {
	    Path: string;
	    Title: string;
	    Description: string;
	    Date: string;
	    Draft: boolean;
	    Tags: string[];
	    Extra: Record<string, string>;
	    BundleType: string;
	
	    static createFrom(source: any = {}) {
	        return new ContentFileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Title = source["Title"];
	        this.Description = source["Description"];
	        this.Date = source["Date"];
	        this.Draft = source["Draft"];
	        this.Tags = source["Tags"];
	        this.Extra = source["Extra"];
	        this.BundleType = source["BundleType"];
	    }
	}
	export class ContentTypeInfo {
	    Name: string;
	    Path: string;
	    FileCount: number;
	    Files: string[];
	
	    static createFrom(source: any = {}) {
	        return new ContentTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Path = source["Path"];
	        this.FileCount = source["FileCount"];
	        this.Files = source["Files"];
	    }
	}
	export class ThemeLayoutInfo {
	    Name: string;
	    SupportedSections: string[];
	    FrontmatterFields: string[];
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeLayoutInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.SupportedSections = source["SupportedSections"];
	        this.FrontmatterFields = source["FrontmatterFields"];
	        this.Description = source["Description"];
	    }
	}
	export class MenuInfo {
	    Name: string;
	    URL: string;
	    Weight: number;
	
	    static createFrom(source: any = {}) {
	        return new MenuInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.URL = source["URL"];
	        this.Weight = source["Weight"];
	    }
	}
	export class SiteConfigInfo {
	    Title: string;
	    BaseURL: string;
	    Language: string;
	    Theme: string;
	    Description: string;
	    Author: string;
	    Menu: MenuInfo[];
	    Params: Record<string, any>;
	    Taxonomies: Record<string, string>;
	    Permalinks: Record<string, string>;
	    Markup: Record<string, any>;
	    Outputs: Record<string, Array<string>>;
	    RawConfig: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new SiteConfigInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Title = source["Title"];
	        this.BaseURL = source["BaseURL"];
	        this.Language = source["Language"];
	        this.Theme = source["Theme"];
	        this.Description = source["Description"];
	        this.Author = source["Author"];
	        this.Menu = this.convertValues(source["Menu"], MenuInfo);
	        this.Params = source["Params"];
	        this.Taxonomies = source["Taxonomies"];
	        this.Permalinks = source["Permalinks"];
	        this.Markup = source["Markup"];
	        this.Outputs = source["Outputs"];
	        this.RawConfig = source["RawConfig"];
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
	export class ContentStructure {
	    SitePath: string;
	    SiteConfig?: SiteConfigInfo;
	    ThemeInfo?: ThemeLayoutInfo;
	    ContentTypes: ContentTypeInfo[];
	    ContentFiles: ContentFileInfo[];
	    DefaultType: string;
	    ContentDir: string;
	
	    static createFrom(source: any = {}) {
	        return new ContentStructure(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SitePath = source["SitePath"];
	        this.SiteConfig = this.convertValues(source["SiteConfig"], SiteConfigInfo);
	        this.ThemeInfo = this.convertValues(source["ThemeInfo"], ThemeLayoutInfo);
	        this.ContentTypes = this.convertValues(source["ContentTypes"], ContentTypeInfo);
	        this.ContentFiles = this.convertValues(source["ContentFiles"], ContentFileInfo);
	        this.DefaultType = source["DefaultType"];
	        this.ContentDir = source["ContentDir"];
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

export namespace api {
	
	export class AIConfigResult {
	    configured: boolean;
	    enabled: boolean;
	    provider?: string;
	    currentProvider?: string;
	    model?: string;
	    currentModel?: string;
	    configuredProviders?: string[];
	    success: boolean;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AIConfigResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configured = source["configured"];
	        this.enabled = source["enabled"];
	        this.provider = source["provider"];
	        this.currentProvider = source["currentProvider"];
	        this.model = source["model"];
	        this.currentModel = source["currentModel"];
	        this.configuredProviders = source["configuredProviders"];
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class AIConfigureParams {
	    provider: string;
	    apiKey: string;
	    baseURL?: string;
	    model?: string;
	
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
	    parentDir?: string;
	    siteName: string;
	    siteType: string;
	    description?: string;
	    audience?: string;
	
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
	    }
	}
	export class AddressListResult {
	    addresses: string[];
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new AddressListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.addresses = source["addresses"];
	        this.error = source["error"];
	    }
	}
	export class ArchiveProjectResult {
	    success: boolean;
	    message: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ArchiveProjectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class ToolVersionInfo {
	    tool: string;
	    currentVersion: string;
	    latestVersion: string;
	    updateRequired: boolean;
	    installed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToolVersionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tool = source["tool"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.updateRequired = source["updateRequired"];
	        this.installed = source["installed"];
	    }
	}
	export class CheckToolVersionsResult {
	    success: boolean;
	    tools: ToolVersionInfo[];
	    message: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new CheckToolVersionsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.tools = this.convertValues(source["tools"], ToolVersionInfo);
	        this.message = source["message"];
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
	export class CreateAddressResult {
	    success: boolean;
	    address: string;
	    alias: string;
	    recoveryPhrase: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateAddressResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.address = source["address"];
	        this.alias = source["alias"];
	        this.recoveryPhrase = source["recoveryPhrase"];
	        this.error = source["error"];
	    }
	}
	export class DeleteProjectParams {
	    projectId: number;
	
	    static createFrom(source: any = {}) {
	        return new DeleteProjectParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	    }
	}
	export class DeleteProjectResult {
	    success: boolean;
	    message: string;
	    error: string;
	    onChainDestroyed: boolean;
	    estimatedGasCost?: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteProjectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.onChainDestroyed = source["onChainDestroyed"];
	        this.estimatedGasCost = source["estimatedGasCost"];
	    }
	}
	export class DeploymentRecord {
	    id: number;
	    projectId: number;
	    objectId: string;
	    network: string;
	    epochs: number;
	    gasFee: string;
	    version?: string;
	    notes?: string;
	    success: boolean;
	    error?: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new DeploymentRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectId = source["projectId"];
	        this.objectId = source["objectId"];
	        this.network = source["network"];
	        this.epochs = source["epochs"];
	        this.gasFee = source["gasFee"];
	        this.version = source["version"];
	        this.notes = source["notes"];
	        this.success = source["success"];
	        this.error = source["error"];
	        this.createdAt = source["createdAt"];
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
	    message: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new EditProjectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class GasEstimateParams {
	    sitePath: string;
	    network: string;
	    epochs: number;
	
	    static createFrom(source: any = {}) {
	        return new GasEstimateParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.network = source["network"];
	        this.epochs = source["epochs"];
	    }
	}
	export class GasEstimateResult {
	    success: boolean;
	    wal: number;
	    sui: number;
	    walRange: string;
	    suiRange: string;
	    summary: string;
	    siteSize: number;
	    fileCount: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new GasEstimateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.wal = source["wal"];
	        this.sui = source["sui"];
	        this.walRange = source["walRange"];
	        this.suiRange = source["suiRange"];
	        this.summary = source["summary"];
	        this.siteSize = source["siteSize"];
	        this.fileCount = source["fileCount"];
	        this.error = source["error"];
	    }
	}
	export class GenerateContentParams {
	    sitePath: string;
	    filePath?: string;
	    contentType: string;
	    topic: string;
	    context: string;
	    instructions: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateContentParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.filePath = source["filePath"];
	        this.contentType = source["contentType"];
	        this.topic = source["topic"];
	        this.context = source["context"];
	        this.instructions = source["instructions"];
	    }
	}
	export class GenerateContentResult {
	    success: boolean;
	    content: string;
	    filePath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateContentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.content = source["content"];
	        this.filePath = source["filePath"];
	        this.error = source["error"];
	    }
	}
	export class GetInstalledThemesResult {
	    success: boolean;
	    themes: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new GetInstalledThemesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.themes = source["themes"];
	        this.error = source["error"];
	    }
	}
	export class ImportAddressParams {
	    method: string;
	    keyScheme: string;
	    input: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportAddressParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.keyScheme = source["keyScheme"];
	        this.input = source["input"];
	    }
	}
	export class ImportAddressResult {
	    success: boolean;
	    address: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportAddressResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.address = source["address"];
	        this.error = source["error"];
	    }
	}
	export class ImportObsidianParams {
	    vaultPath: string;
	    siteName: string;
	    parentDir: string;
	    outputDir: string;
	    dryRun: boolean;
	    convertLinks: boolean;
	    linkStyle: string;
	    includeDrafts: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ImportObsidianParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vaultPath = source["vaultPath"];
	        this.siteName = source["siteName"];
	        this.parentDir = source["parentDir"];
	        this.outputDir = source["outputDir"];
	        this.dryRun = source["dryRun"];
	        this.convertLinks = source["convertLinks"];
	        this.linkStyle = source["linkStyle"];
	        this.includeDrafts = source["includeDrafts"];
	    }
	}
	export class ImportObsidianResult {
	    success: boolean;
	    filesImported: number;
	    sitePath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportObsidianResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.filesImported = source["filesImported"];
	        this.sitePath = source["sitePath"];
	        this.error = source["error"];
	    }
	}
	export class InitSiteResult {
	    success: boolean;
	    sitePath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new InitSiteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.sitePath = source["sitePath"];
	        this.error = source["error"];
	    }
	}
	export class InstallThemeParams {
	    sitePath: string;
	    githubUrl: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallThemeParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.githubUrl = source["githubUrl"];
	    }
	}
	export class InstallThemeResult {
	    success: boolean;
	    themeName: string;
	    removedThemes?: string[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallThemeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.themeName = source["themeName"];
	        this.removedThemes = source["removedThemes"];
	        this.error = source["error"];
	    }
	}
	export class LaunchStep {
	    name: string;
	    status: string;
	    message: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new LaunchStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class LaunchWizardParams {
	    sitePath?: string;
	    network: string;
	    projectName: string;
	    category: string;
	    description: string;
	    imageUrl: string;
	    epochs: number;
	    skipConfirm: boolean;
	
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
	        this.imageUrl = source["imageUrl"];
	        this.epochs = source["epochs"];
	        this.skipConfirm = source["skipConfirm"];
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
	    sitePath?: string;
	    slug: string;
	    contentType: string;
	    noBuild: boolean;
	    serve: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NewContentParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sitePath = source["sitePath"];
	        this.slug = source["slug"];
	        this.contentType = source["contentType"];
	        this.noBuild = source["noBuild"];
	        this.serve = source["serve"];
	    }
	}
	export class NewContentResult {
	    success: boolean;
	    path: string;
	    filePath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new NewContentResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.path = source["path"];
	        this.filePath = source["filePath"];
	        this.error = source["error"];
	    }
	}
	export class Project {
	    id: number;
	    name: string;
	    description: string;
	    category: string;
	    objectId: string;
	    network: string;
	    wallet: string;
	    sitePath: string;
	    imageUrl: string;
	    suins: string;
	    createdAt: string;
	    updatedAt: string;
	    lastDeployAt: string;
	    firstDeployAt?: string;
	    epochs: number;
	    totalEpochs?: number;
	    gasFee?: string;
	    expiresAt?: string;
	    expiresIn?: string;
	    status: string;
	    deployCount: number;
	    size?: number;
	    fileCount?: number;
	    deployments?: DeploymentRecord[];
	    suiReady?: boolean;
	    walrusReady?: boolean;
	    siteBuilder?: boolean;
	    hugoReady?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.category = source["category"];
	        this.objectId = source["objectId"];
	        this.network = source["network"];
	        this.wallet = source["wallet"];
	        this.sitePath = source["sitePath"];
	        this.imageUrl = source["imageUrl"];
	        this.suins = source["suins"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.lastDeployAt = source["lastDeployAt"];
	        this.firstDeployAt = source["firstDeployAt"];
	        this.epochs = source["epochs"];
	        this.totalEpochs = source["totalEpochs"];
	        this.gasFee = source["gasFee"];
	        this.expiresAt = source["expiresAt"];
	        this.expiresIn = source["expiresIn"];
	        this.status = source["status"];
	        this.deployCount = source["deployCount"];
	        this.size = source["size"];
	        this.fileCount = source["fileCount"];
	        this.deployments = this.convertValues(source["deployments"], DeploymentRecord);
	        this.suiReady = source["suiReady"];
	        this.walrusReady = source["walrusReady"];
	        this.siteBuilder = source["siteBuilder"];
	        this.hugoReady = source["hugoReady"];
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
	export class ProviderCredentialsResult {
	    success: boolean;
	    apiKey: string;
	    baseURL: string;
	    model: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderCredentialsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.apiKey = source["apiKey"];
	        this.baseURL = source["baseURL"];
	        this.model = source["model"];
	        this.error = source["error"];
	    }
	}
	export class QuickStartParams {
	    parentDir?: string;
	    siteName: string;
	    name?: string;
	    siteType?: string;
	    skipBuild: boolean;
	
	    static createFrom(source: any = {}) {
	        return new QuickStartParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentDir = source["parentDir"];
	        this.siteName = source["siteName"];
	        this.name = source["name"];
	        this.siteType = source["siteType"];
	        this.skipBuild = source["skipBuild"];
	    }
	}
	export class QuickStartResult {
	    success: boolean;
	    sitePath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new QuickStartResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.sitePath = source["sitePath"];
	        this.error = source["error"];
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
	export class SetStatusParams {
	    projectId: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new SetStatusParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.status = source["status"];
	    }
	}
	export class SetStatusResult {
	    success: boolean;
	    message: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SetStatusResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class SetupDepsResult {
	    success: boolean;
	    message: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SetupDepsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class SwitchAddressResult {
	    success: boolean;
	    address: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SwitchAddressResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.address = source["address"];
	        this.error = source["error"];
	    }
	}
	export class SwitchNetworkResult {
	    success: boolean;
	    network: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new SwitchNetworkResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.network = source["network"];
	        this.error = source["error"];
	    }
	}
	export class SystemHealth {
	    netOnline: boolean;
	    suiInstalled: boolean;
	    suiConfigured: boolean;
	    walrusInstalled: boolean;
	    siteBuilder: boolean;
	    hugoInstalled: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.netOnline = source["netOnline"];
	        this.suiInstalled = source["suiInstalled"];
	        this.suiConfigured = source["suiConfigured"];
	        this.walrusInstalled = source["walrusInstalled"];
	        this.siteBuilder = source["siteBuilder"];
	        this.hugoInstalled = source["hugoInstalled"];
	        this.message = source["message"];
	    }
	}
	
	export class UpdateContentParams {
	    filePath: string;
	    instructions: string;
	    sitePath: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateContentParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.instructions = source["instructions"];
	        this.sitePath = source["sitePath"];
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
	export class UpdateSiteParams {
	    projectId: number;
	    epochs?: number;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSiteParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectId = source["projectId"];
	        this.epochs = source["epochs"];
	    }
	}
	export class UpdateSiteResult {
	    success: boolean;
	    objectId: string;
	    gasFee: string;
	    message: string;
	    error: string;
	    logs?: string[];
	
	    static createFrom(source: any = {}) {
	        return new UpdateSiteResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.objectId = source["objectId"];
	        this.gasFee = source["gasFee"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.logs = source["logs"];
	    }
	}
	export class UpdateToolsParams {
	    tools: string[];
	    network: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateToolsParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tools = source["tools"];
	        this.network = source["network"];
	    }
	}
	export class UpdateToolsResult {
	    success: boolean;
	    updatedTools: string[];
	    failedTools: Record<string, string>;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateToolsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.updatedTools = source["updatedTools"];
	        this.failedTools = source["failedTools"];
	        this.message = source["message"];
	    }
	}
	export class VersionResult {
	    version: string;
	    gitCommit: string;
	    buildDate: string;
	
	    static createFrom(source: any = {}) {
	        return new VersionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.gitCommit = source["gitCommit"];
	        this.buildDate = source["buildDate"];
	    }
	}
	export class WalletInfo {
	    address: string;
	    suiBalance: number;
	    walBalance: number;
	    network: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WalletInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.suiBalance = source["suiBalance"];
	        this.walBalance = source["walBalance"];
	        this.network = source["network"];
	        this.active = source["active"];
	    }
	}

}

export namespace main {
	
	export class AIProgressState {
	    isActive: boolean;
	    phase: string;
	    message: string;
	    pagePath: string;
	    progress: number;
	    current: number;
	    total: number;
	    complete: boolean;
	    success: boolean;
	    sitePath: string;
	    totalPages: number;
	    filesCreated: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new AIProgressState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isActive = source["isActive"];
	        this.phase = source["phase"];
	        this.message = source["message"];
	        this.pagePath = source["pagePath"];
	        this.progress = source["progress"];
	        this.current = source["current"];
	        this.total = source["total"];
	        this.complete = source["complete"];
	        this.success = source["success"];
	        this.sitePath = source["sitePath"];
	        this.totalPages = source["totalPages"];
	        this.filesCreated = source["filesCreated"];
	        this.error = source["error"];
	    }
	}
	export class CheckDirectoryDepthResult {
	    success: boolean;
	    depth: number;
	    tooDeep: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new CheckDirectoryDepthResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.depth = source["depth"];
	        this.tooDeep = source["tooDeep"];
	        this.error = source["error"];
	    }
	}
	export class CopyFileResult {
	    success: boolean;
	    error: string;
	    srcPath: string;
	    dstPath: string;
	
	    static createFrom(source: any = {}) {
	        return new CopyFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.srcPath = source["srcPath"];
	        this.dstPath = source["dstPath"];
	    }
	}
	export class CreateDirectoryResult {
	    success: boolean;
	    path: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateDirectoryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.path = source["path"];
	        this.error = source["error"];
	    }
	}
	export class CreateFileResult {
	    success: boolean;
	    path: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.path = source["path"];
	        this.error = source["error"];
	    }
	}
	export class DeleteFileResult {
	    success: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}
	export class FileInfo {
	    name: string;
	    path: string;
	    isDir: boolean;
	    size: number;
	    modified: number;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.isDir = source["isDir"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	    }
	}
	export class FolderStatsResult {
	    success: boolean;
	    fileCount: number;
	    folderCount: number;
	    totalSize: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new FolderStatsResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.fileCount = source["fileCount"];
	        this.folderCount = source["folderCount"];
	        this.totalSize = source["totalSize"];
	        this.error = source["error"];
	    }
	}
	export class ListFilesResult {
	    success: boolean;
	    files: FileInfo[];
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ListFilesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.files = this.convertValues(source["files"], FileInfo);
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
	export class MoveFileResult {
	    success: boolean;
	    error: string;
	    oldPath: string;
	    newPath: string;
	
	    static createFrom(source: any = {}) {
	        return new MoveFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.oldPath = source["oldPath"];
	        this.newPath = source["newPath"];
	    }
	}
	export class ReadFileResult {
	    success: boolean;
	    content: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ReadFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.content = source["content"];
	        this.error = source["error"];
	    }
	}
	export class RenameFileResult {
	    success: boolean;
	    error: string;
	    oldPath: string;
	    newPath: string;
	
	    static createFrom(source: any = {}) {
	        return new RenameFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.oldPath = source["oldPath"];
	        this.newPath = source["newPath"];
	    }
	}
	export class WriteFileResult {
	    success: boolean;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new WriteFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	    }
	}

}

