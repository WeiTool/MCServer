export namespace model {
	
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

}

