export namespace deals {
	
	export class Listing {
	    id: string;
	    title: string;
	    price: number;
	    currency: string;
	    condition: string;
	    url: string;
	    imageUrl: string;
	    seller: string;
	    location: string;
	    source: string;
	    // Go type: time
	    listedAt: any;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new Listing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.price = source["price"];
	        this.currency = source["currency"];
	        this.condition = source["condition"];
	        this.url = source["url"];
	        this.imageUrl = source["imageUrl"];
	        this.seller = source["seller"];
	        this.location = source["location"];
	        this.source = source["source"];
	        this.listedAt = this.convertValues(source["listedAt"], null);
	        this.score = source["score"];
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
	export class SearchResult {
	    listings: Listing[];
	    ebayError?: string;
	    facebookError?: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.listings = this.convertValues(source["listings"], Listing);
	        this.ebayError = source["ebayError"];
	        this.facebookError = source["facebookError"];
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

export namespace drive {
	
	export class DriveInfo {
	    path: string;
	    label: string;
	    sizeBytes: number;
	    isRemovable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DriveInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.label = source["label"];
	        this.sizeBytes = source["sizeBytes"];
	        this.isRemovable = source["isRemovable"];
	    }
	}

}

export namespace library {
	
	export class Game {
	    path: string;
	    title: string;
	    gameId: string;
	    region: string;
	    sizeBytes: number;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Game(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.title = source["title"];
	        this.gameId = source["gameId"];
	        this.region = source["region"];
	        this.sizeBytes = source["sizeBytes"];
	        this.type = source["type"];
	    }
	}

}

