export namespace deps {
	
	export class DependencyState {
	    name: string;
	    current_version: string;
	    last_checked: number;
	    status: string;
	    needs_update: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DependencyState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.current_version = source["current_version"];
	        this.last_checked = source["last_checked"];
	        this.status = source["status"];
	        this.needs_update = source["needs_update"];
	    }
	}

}

export namespace domain {
	
	export class Cookies {
	    is_allowed: boolean;
	    path?: string;
	    browser?: string;
	    type?: string;
	
	    static createFrom(source: any = {}) {
	        return new Cookies(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_allowed = source["is_allowed"];
	        this.path = source["path"];
	        this.browser = source["browser"];
	        this.type = source["type"];
	    }
	}
	export class DownloadProgress {
	    percentage: number;
	    downloaded_bytes: number;
	    logs: string[];
	
	    static createFrom(source: any = {}) {
	        return new DownloadProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.percentage = source["percentage"];
	        this.downloaded_bytes = source["downloaded_bytes"];
	        this.logs = source["logs"];
	    }
	}
	export class Subtitle {
	    is_allowed: boolean;
	    language_codes?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Subtitle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_allowed = source["is_allowed"];
	        this.language_codes = source["language_codes"];
	    }
	}
	export class TimeRange {
	    is_allowed: boolean;
	    start?: string;
	    end?: string;
	
	    static createFrom(source: any = {}) {
	        return new TimeRange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.is_allowed = source["is_allowed"];
	        this.start = source["start"];
	        this.end = source["end"];
	    }
	}
	export class PlaylistSelection {
	    type: string;
	    start_index: number;
	    end_index: number;
	    items: string;
	
	    static createFrom(source: any = {}) {
	        return new PlaylistSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.start_index = source["start_index"];
	        this.end_index = source["end_index"];
	        this.items = source["items"];
	    }
	}
	export class Media {
	    id: string;
	    title: string;
	    total_bytes: number;
	    url: string;
	    file_path: string;
	    quality: number;
	    only_audio: boolean;
	    cookies?: Cookies;
	    status: number;
	    progress: DownloadProgress;
	    is_playlist: boolean;
	    playlist_selection?: PlaylistSelection;
	    time_range?: TimeRange;
	    subtitle?: Subtitle;
	
	    static createFrom(source: any = {}) {
	        return new Media(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.total_bytes = source["total_bytes"];
	        this.url = source["url"];
	        this.file_path = source["file_path"];
	        this.quality = source["quality"];
	        this.only_audio = source["only_audio"];
	        this.cookies = this.convertValues(source["cookies"], Cookies);
	        this.status = source["status"];
	        this.progress = this.convertValues(source["progress"], DownloadProgress);
	        this.is_playlist = source["is_playlist"];
	        this.playlist_selection = this.convertValues(source["playlist_selection"], PlaylistSelection);
	        this.time_range = this.convertValues(source["time_range"], TimeRange);
	        this.subtitle = this.convertValues(source["subtitle"], Subtitle);
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
	export class MediaDefaults {
	    quality: number;
	    download_path: string;
	    only_audio: boolean;
	    cookies?: Cookies;
	    subtitle?: Subtitle;
	
	    static createFrom(source: any = {}) {
	        return new MediaDefaults(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quality = source["quality"];
	        this.download_path = source["download_path"];
	        this.only_audio = source["only_audio"];
	        this.cookies = this.convertValues(source["cookies"], Cookies);
	        this.subtitle = this.convertValues(source["subtitle"], Subtitle);
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
	
	export class Setting {
	    parallel_downloads: number;
	
	    static createFrom(source: any = {}) {
	        return new Setting(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parallel_downloads = source["parallel_downloads"];
	    }
	}
	

}

export namespace updater {
	
	export class UpdateResult {
	    success: boolean;
	    message: string;
	    current_version?: string;
	    latest_version?: string;
	    has_update?: boolean;
	    changelog?: string;
	    download_url?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.current_version = source["current_version"];
	        this.latest_version = source["latest_version"];
	        this.has_update = source["has_update"];
	        this.changelog = source["changelog"];
	        this.download_url = source["download_url"];
	    }
	}

}

