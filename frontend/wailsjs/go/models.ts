export namespace livesearch {
	
	export class TradeLink {
	    id: number;
	    url: string;
	    description: string;
	    selected: boolean;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new TradeLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.url = source["url"];
	        this.description = source["description"];
	        this.selected = source["selected"];
	        this.status = source["status"];
	    }
	}

}

export namespace settings {
	
	export class Config {
	    poesessid: string;
	    accountName: string;
	    league: string;
	    automationEnabled: boolean;
	    delay: number;
	
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
	    }
	}

}

