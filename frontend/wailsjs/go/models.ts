export namespace mrat {
	
	export class UGenArg {
	    name: string;
	    default: any;
	    doc: string;
	
	    static createFrom(source: any = {}) {
	        return new UGenArg(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.default = source["default"];
	        this.doc = source["doc"];
	    }
	}
	export class Symbol {
	    name: string;
	    group: string;
	    doc: string;
	    arglists: any[];
	    ugenargs: UGenArg[];
	
	    static createFrom(source: any = {}) {
	        return new Symbol(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.group = source["group"];
	        this.doc = source["doc"];
	        this.arglists = source["arglists"];
	        this.ugenargs = this.convertValues(source["ugenargs"], UGenArg);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

