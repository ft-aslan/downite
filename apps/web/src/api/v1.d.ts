/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */

export interface paths {
    "/download": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get all downloads */
        get: operations["get-downloads"];
        put?: never;
        /** Download with url */
        post: operations["download"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/download/delete": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Delete download with files */
        post: operations["delete-download-with-files"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/download/meta": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Get meta data of download */
        post: operations["get-download-meta"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/download/pause": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Pause download */
        post: operations["pause-download"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/download/remove": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Remove download */
        post: operations["remove-download"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/download/resume": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Resume download */
        post: operations["resume-download"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/download/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get download */
        get: operations["get-download"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/meta/file": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Get torrent meta info with file */
        post: operations["get-torrent-meta-info-with-file"];
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
        /** Get torrent meta info with magnet */
        post: operations["get-torrent-meta-info-with-magnet"];
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
        /** Download torrent */
        post: operations["download-torrent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent/delete": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Delete torrent */
        post: operations["delete-torrent"];
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
        /** Pause torrent */
        post: operations["pause-torrent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent/remove": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Remove torrent */
        post: operations["remove-torrent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent/resume": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Resume torrent */
        post: operations["resume-torrent"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent/speed": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get torrents total speed */
        get: operations["get-torrents-total-speed"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/torrent/{infohash}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get torrent */
        get: operations["get-torrent"];
        put?: never;
        post?: never;
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
        Download: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            /** Format: int64 */
            DownloadedBytes: number;
            /** Format: int64 */
            PartCount: number;
            /** Format: int64 */
            PartLength: number;
            /** Format: int64 */
            QueueNumber: number;
            /** Format: int64 */
            TotalSize: number;
            /** Format: date-time */
            createdAt: string;
            /** Format: date-time */
            finishedAt: string;
            /** Format: int64 */
            id: number;
            name: string;
            parts: components["schemas"]["DownloadPart"][];
            path: string;
            /** Format: date-time */
            startedAt: string;
            /** Format: int64 */
            status: number;
            /** Format: int64 */
            timeActive: number;
            url: string;
        };
        DownloadActionReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            ids: number[];
        };
        DownloadActionResBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            success: boolean;
        };
        DownloadMeta: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            fileName: string;
            fileType: string;
            isRangeAllowed: boolean;
            /** Format: int64 */
            totalSize: number;
            url: string;
        };
        DownloadPart: {
            /** Format: date-time */
            createdAt: string;
            /** Format: int64 */
            downloadedBytes: number;
            /** Format: int64 */
            endByteIndex: number;
            /** Format: date-time */
            finishedAt: string;
            /** Format: int64 */
            partIndex: number;
            /** Format: int64 */
            partLength: number;
            /** Format: int64 */
            startByteIndex: number;
            /** Format: date-time */
            startedAt: string;
            /** Format: int64 */
            status: number;
            /** Format: int64 */
            timeActive: number;
        };
        DownloadReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            addTopOfQueue: boolean;
            category: string;
            /** @enum {string} */
            contentLayout: "Original" | "Create subfolder" | "Don't create subfolder";
            description: string;
            incompleteSavePath: string;
            isIncompleteSavePathEnabled: boolean;
            savePath: string;
            startDownload: boolean;
            tags: string[];
            url: string;
        };
        DownloadTorrentReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            addTopOfQueue: boolean;
            category?: string;
            /** @enum {string} */
            contentLayout: "Original" | "Create subfolder" | "Don't create subfolder";
            downloadSequentially: boolean;
            files: components["schemas"]["TorrentFileFlatTreeNode"][];
            incompleteSavePath?: string;
            isIncompleteSavePathEnabled: boolean;
            magnet: string;
            savePath: string;
            skipHashCheck: boolean;
            startTorrent: boolean;
            tags?: string[];
            torrentFile?: unknown;
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
        GetDownloadMetaReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            url: string;
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
        Peer: {
            url: string;
        };
        PieceProgress: {
            /** Format: int64 */
            downloadedByteCount: number;
            /** Format: int64 */
            index: number;
            /** Format: int64 */
            length: number;
        };
        Torrent: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            /** Format: int64 */
            amountLeft: number;
            /** Format: float */
            availability: number;
            category: string;
            comment: string;
            /** Format: int64 */
            createdAt: number;
            /** Format: float */
            downloadSpeed: number;
            /** Format: int64 */
            downloaded: number;
            /** Format: int64 */
            eta: number;
            files: components["schemas"]["TorrentFileTreeNode"][];
            infohash: string;
            magnet: string;
            name: string;
            /** Format: int64 */
            peerCount: number;
            peers: components["schemas"]["Peer"][];
            pieceProgress: components["schemas"]["PieceProgress"][];
            /** Format: float */
            progress: number;
            /** Format: int64 */
            queueNumber: number;
            /** Format: float */
            ratio: number;
            savePath: string;
            /** Format: int64 */
            seeds: number;
            /** Format: int64 */
            sizeOfWanted: number;
            /** Format: int64 */
            startedAt: number;
            /** @enum {string} */
            status: "paused" | "downloading" | "completed" | "seeding" | "metadata";
            tags: string[];
            /** Format: int64 */
            timeActive: number;
            /** Format: int64 */
            totalSize: number;
            trackers: components["schemas"]["Tracker"][];
            /** Format: float */
            uploadSpeed: number;
            /** Format: int64 */
            uploaded: number;
        };
        TorrentActionReqBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            /** @description Hashes of torrents */
            infoHashes: string[];
        };
        TorrentActionResBody: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            result: boolean;
        };
        TorrentFileFlatTreeNode: {
            name: string;
            path: string;
            /** @enum {string} */
            priority: "none" | "low" | "normal" | "high" | "maximum";
        };
        TorrentFileTreeNode: {
            children: components["schemas"]["TorrentFileTreeNode"][];
            /** Format: int64 */
            length: number;
            name: string;
            path: string;
            /** @enum {string} */
            priority: "none" | "low" | "normal" | "high" | "maximum";
        };
        TorrentMeta: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            files: components["schemas"]["TorrentFileTreeNode"][];
            infohash: string;
            magnet: string;
            name: string;
            /** Format: int64 */
            totalSize: number;
        };
        TorrentsTotalSpeedData: {
            /**
             * Format: uri
             * @description A URL to the JSON Schema for this object.
             */
            readonly $schema?: string;
            /** Format: float */
            downloadSpeed: number;
            time: string;
            /** Format: float */
            uploadSpeed: number;
        };
        Tracker: {
            /** Format: int64 */
            interval: number;
            peers: components["schemas"]["Peer"][];
            /** Format: int64 */
            tier: number;
            url: string;
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
    "get-downloads": {
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
                    "application/json": components["schemas"]["Download"][];
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
    download: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DownloadReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Download"];
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
    "delete-download-with-files": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DownloadActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DownloadActionResBody"];
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
    "get-download-meta": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["GetDownloadMetaReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DownloadMeta"];
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
    "pause-download": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DownloadActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DownloadActionResBody"];
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
    "remove-download": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DownloadActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DownloadActionResBody"];
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
    "resume-download": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["DownloadActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DownloadActionResBody"];
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
    "get-download": {
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
                    "application/json": components["schemas"]["Download"];
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
    "get-torrent-meta-info-with-file": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "multipart/form-data": Record<string, never>;
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
    "get-torrent-meta-info-with-magnet": {
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
    "download-torrent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "multipart/form-data": components["schemas"]["DownloadTorrentReqBody"];
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
    "delete-torrent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["TorrentActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TorrentActionResBody"];
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
    "pause-torrent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["TorrentActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TorrentActionResBody"];
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
    "remove-torrent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["TorrentActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TorrentActionResBody"];
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
    "resume-torrent": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["TorrentActionReqBody"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TorrentActionResBody"];
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
    "get-torrents-total-speed": {
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
                    "application/json": components["schemas"]["TorrentsTotalSpeedData"];
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
    "get-torrent": {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /**
                 * @description Infohash of the torrent
                 * @example 2b66980093bc11806fab50cb3cb41835b95a0362
                 */
                infohash: string;
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
}
