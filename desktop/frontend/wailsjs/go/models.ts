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

}

