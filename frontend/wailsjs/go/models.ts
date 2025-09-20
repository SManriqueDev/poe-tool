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

export namespace logging {
	
	export class LogConfig {
	    enabled: boolean;
	    log_level: string;
	    log_modules: string[];
	    log_new_items: boolean;
	    log_api_requests: boolean;
	    log_websocket: boolean;
	    retention_days: number;
	    max_entries: number;
	    real_time_updates: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LogConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.log_level = source["log_level"];
	        this.log_modules = source["log_modules"];
	        this.log_new_items = source["log_new_items"];
	        this.log_api_requests = source["log_api_requests"];
	        this.log_websocket = source["log_websocket"];
	        this.retention_days = source["retention_days"];
	        this.max_entries = source["max_entries"];
	        this.real_time_updates = source["real_time_updates"];
	    }
	}
	export class LogEntry {
	    id: number;
	    timestamp: time.Time;
	    module: string;
	    level: string;
	    message: string;
	    metadata: string;
	    created_at: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], time.Time);
	        this.module = source["module"];
	        this.level = source["level"];
	        this.message = source["message"];
	        this.metadata = source["metadata"];
	        this.created_at = this.convertValues(source["created_at"], time.Time);
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
	export class LogFilter {
	    module?: string;
	    level?: string;
	    start_time?: time.Time;
	    end_time?: time.Time;
	    search?: string;
	    limit?: number;
	    offset?: number;
	
	    static createFrom(source: any = {}) {
	        return new LogFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.module = source["module"];
	        this.level = source["level"];
	        this.start_time = this.convertValues(source["start_time"], time.Time);
	        this.end_time = this.convertValues(source["end_time"], time.Time);
	        this.search = source["search"];
	        this.limit = source["limit"];
	        this.offset = source["offset"];
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

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

