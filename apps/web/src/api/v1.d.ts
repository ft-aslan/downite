/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */

export interface paths {
    "/meta/file": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Post meta file */
        post: operations["post-meta-file"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/meta/magnet": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Post meta magnet */
        post: operations["post-meta-magnet"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get all torrents */
        get: operations["get-all-torrents"];
        put?: never;
        /** Post torrent */
        post: operations["post-torrent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent/:hash": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get torrent hash */
        get: operations["get-torrent-hash"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent/pause": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Post torrent pause */
        post: operations["post-torrent-pause"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        DownloadTorrentReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            addTopOfQueue: boolean;
            category?: string;
            contentLayout: string;
            downloadSequentially: boolean;
            files: components["schemas"]["TorrentFileOptions"][];
            incompleteSavePath?: string;
            isIncompleteSavePathEnabled: boolean;
            magnet?: string;
            savePath: string;
            skipHashCheck: boolean;
            startTorrent: boolean;
            tags?: string[];
            torrentFile?: string;
        };
        ErrorDetail: {
            /** @description Where the error occurred, e.g. 'body.items[3].tags' or 'path.thing-id' */
            location?: string;
            /** @description Error message text */
            message?: string;
            /** @description The value at the given location */
            value?: unknown;
        };
        ErrorModel: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            /** @description A human-readable explanation specific to this occurrence of the problem. */
            detail?: string;
            /** @description Optional list of individual error details */
            errors?: components["schemas"]["ErrorDetail"][];
            /**
             * Format: uri
             * @description A URI reference that identifies the specific occurrence of the problem.
             */
            instance?: string;
            /**
             * Format: int64
             * @description HTTP status code
             */
            status?: number;
            /** @description A short, human-readable summary of the problem type. This value should not change between occurrences of the error. */
            title?: string;
            /**
             * Format: uri
             * @description A URI reference to human-readable documentation for the error.
             * @default about:blank
             */
            type: string;
        };
        FileTree: {
            Dir: {
                [key: string]: components["schemas"]["FileTree"] | undefined;
            };
            File: components["schemas"]["FileTreeFileStruct"];
        };
        FileTreeFileStruct: {
            /** Format: int64 */
            Length: number;
            PiecesRoot: string;
        };
        GetMetaWithMagnetReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            magnet: string;
        };
        GetTorrentsResBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            torrents: components["schemas"]["Torrent"][];
        };
        PauseTorrentReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            /** @description Hash of the torrent */
            hashes: string[];
        };
        PauseTorrentResBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            result: boolean;
        };
        PeerInfo: {
            Addr: unknown;
            Id: string;
            Source: string;
            SupportsEncryption: boolean;
            Trusted: boolean;
        };
        PieceProgress: {
            /** Format: int64 */
            DownloadedByteCount: number;
            /** Format: int64 */
            Index: number;
            /** Format: int64 */
            Length: number;
        };
        Torrent: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            /** Format: int64 */
            addedOn: number;
            /** Format: int64 */
            amountLeft: number;
            /** Format: float */
            availability: number;
            category: string;
            downloadDir: string;
            downloadPath: string;
            /** Format: int64 */
            downloadSpeed: number;
            /** Format: int64 */
            downloaded: number;
            /** Format: int64 */
            eta: number;
            files: components["schemas"]["FileTree"];
            infoHash: string;
            magnet: string;
            name: string;
            peers: {
                [key: string]: components["schemas"]["PeerInfo"] | undefined;
            };
            /** Format: int64 */
            peersCount: number;
            pieceProgress: components["schemas"]["PieceProgress"][];
            /** Format: float */
            progress: number;
            /** Format: float */
            ratio: number;
            /** Format: int64 */
            seeds: number;
            /** Format: int64 */
            status: number;
            tags: string[];
            /** Format: int64 */
            totalSize: number;
            /** Format: int64 */
            uploadSpeed: number;
            /** Format: int64 */
            uploaded: number;
        };
        TorrentFileOptions: {
            /** @enum {string} */
            downloadPriority: "None" | "Low" | "Normal" | "High" | "Maximum";
            name: string;
            path: string;
        };
        TorrentMeta: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            files: components["schemas"]["TreeNodeMeta"][];
            infoHash: string;
            name: string;
            torrentMagnet: string;
            /** Format: int64 */
            totalSize: number;
        };
        TreeNodeMeta: {
            children: components["schemas"]["TreeNodeMeta"][];
            /** Format: int64 */
            length: number;
            name: string;
            path: string[];
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    "post-meta-file": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "multipart/form-data": {
                    /**
                     * Format: binary
                     * @description filename of the file being uploaded
                     */
                    filename?: string;
                    /** @description general purpose name for multipart form value */
                    name?: string;
                };
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TorrentMeta"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/problem+json": components["schemas"]["ErrorModel"];
                };
            };
        };
    };
    "post-meta-magnet": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["GetMetaWithMagnetReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TorrentMeta"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/problem+json": components["schemas"]["ErrorModel"];
                };
            };
        };
    };
    "get-all-torrents": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["GetTorrentsResBody"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/problem+json": components["schemas"]["ErrorModel"];
                };
            };
        };
    };
    "post-torrent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DownloadTorrentReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Torrent"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/problem+json": components["schemas"]["ErrorModel"];
                };
            };
        };
    };
    "get-torrent-hash": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @example 2b66980093bc11806fab50cb3cb41835b95a0362 */
                hash: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Torrent"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/problem+json": components["schemas"]["ErrorModel"];
                };
            };
        };
    };
    "post-torrent-pause": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["PauseTorrentReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PauseTorrentResBody"];
                };
            };
            /** @description Error */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/problem+json": components["schemas"]["ErrorModel"];
                };
            };
        };
    };
}
