export namespace model {
	
	export class FileOperationResponse {
	    success: boolean;
	    message: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new FileOperationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.path = source["path"];
	    }
	}
	export class GlobalConfig {
	    previewEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GlobalConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.previewEnabled = source["previewEnabled"];
	    }
	}
	export class JavaInfo {
	    path: string;
	    executable: string;
	    version: number;
	    versionName: string;
	
	    static createFrom(source: any = {}) {
	        return new JavaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.executable = source["executable"];
	        this.version = source["version"];
	        this.versionName = source["versionName"];
	    }
	}
	export class ServerInstance {
	    name: string;
	    path: string;
	    hasJar: boolean;
	    jarCount: number;
	    jarFiles: string[];
	    hasEula: boolean;
	    eulaAgreed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ServerInstance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.hasJar = source["hasJar"];
	        this.jarCount = source["jarCount"];
	        this.jarFiles = source["jarFiles"];
	        this.hasEula = source["hasEula"];
	        this.eulaAgreed = source["eulaAgreed"];
	    }
	}
	export class ServerListResult {
	    servers: ServerInstance[];
	    basePath: string;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new ServerListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.servers = this.convertValues(source["servers"], ServerInstance);
	        this.basePath = source["basePath"];
	        this.total = source["total"];
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
	export class UpdateState {
	    status: string;
	    version: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.version = source["version"];
	        this.error = source["error"];
	    }
	}
	export class VersionResponse {
	    hasUpdate: boolean;
	    current: string;
	    latest: string;
	    assetName: string;
	    downloadUrl: string;
	    isBeta: boolean;
	    isPreview: boolean;
	
	    static createFrom(source: any = {}) {
	        return new VersionResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasUpdate = source["hasUpdate"];
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.assetName = source["assetName"];
	        this.downloadUrl = source["downloadUrl"];
	        this.isBeta = source["isBeta"];
	        this.isPreview = source["isPreview"];
	    }
	}

}

