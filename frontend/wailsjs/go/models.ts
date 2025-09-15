export namespace livesearch {
	
	export class TradeLink {
	    league: string;
	    searchId: string;
	    url: string;
	    description: string;
	    selected: boolean;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new TradeLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.league = source["league"];
	        this.searchId = source["searchId"];
	        this.url = source["url"];
	        this.description = source["description"];
	        this.selected = source["selected"];
	        this.status = source["status"];
	    }
	}

}

export namespace settings {
	
	export class DefaultTradeLink {
	    url: string;
	    description: string;
	    selected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DefaultTradeLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.description = source["description"];
	        this.selected = source["selected"];
	    }
	}
	export class Config {
	    poesessid: string;
	    accountName: string;
	    league: string;
	    automationEnabled: boolean;
	    delay: number;
	    defaultTradeLinks: DefaultTradeLink[];
	    goToHideout: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.poesessid = source["poesessid"];
	        this.accountName = source["accountName"];
	        this.league = source["league"];
	        this.automationEnabled = source["automationEnabled"];
	        this.delay = source["delay"];
	        this.defaultTradeLinks = this.convertValues(source["defaultTradeLinks"], DefaultTradeLink);
	        this.goToHideout = source["goToHideout"];
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

