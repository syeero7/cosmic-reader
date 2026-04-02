export namespace main {
	
	export class ComicInfo {
	    pageCount: number;
	    title: string;
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new ComicInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pageCount = source["pageCount"];
	        this.title = source["title"];
	        this.id = source["id"];
	    }
	}

}

