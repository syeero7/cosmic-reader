export namespace main {
	
	export class Archive {
	    pageCount: number;
	    title: string;
	    thumbnail: string;
	
	    static createFrom(source: any = {}) {
	        return new Archive(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pageCount = source["pageCount"];
	        this.title = source["title"];
	        this.thumbnail = source["thumbnail"];
	    }
	}

}

