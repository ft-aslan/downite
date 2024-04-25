/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */


export interface paths {
  "/torrent": {
    /** Get all torrents */
    get: operations["get-all-torrents"];
    /** Post torrent */
    post: operations["post-torrent"];
  };
  "/torrent-meta": {
    /** Post torrent meta */
    post: operations["post-torrent-meta"];
  };
  "/torrent/:hash": {
    /** Get torrent hash */
    get: operations["get-torrent-hash"];
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
      $schema?: string;
      addTopOfQueue: boolean;
      category?: string;
      contentLayout: string;
      downloadSequentially: boolean;
      files: components["schemas"]["TorrentFileOptions"][];
      incompleteSavePath: string;
      isIncompleteSavePathEnabled: boolean;
      magnet: string;
      savePath: string;
      skipHashCheck: boolean;
      startTorrent: boolean;
      tags?: string[];
      torrentFile: string;
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
      $schema?: string;
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
      type?: string;
    };
    FileMeta: {
      children: components["schemas"]["FileMeta"][];
      /** Format: int64 */
      length: number;
      name: string;
      path: string[];
    };
    FileTree: {
      Dir: {
        [key: string]: components["schemas"]["FileTree"];
      };
      File: components["schemas"]["FileTreeFileStruct"];
    };
    FileTreeFileStruct: {
      /** Format: int64 */
      Length: number;
      PiecesRoot: string;
    };
    GetTorrentMetaReqBody: {
      /**
       * Format: uri
       * @description A URL to the JSON Schema for this object.
       */
      $schema?: string;
      magnet?: string;
      torrentFile?: string;
    };
    GetTorrentsResBody: {
      /**
       * Format: uri
       * @description A URL to the JSON Schema for this object.
       */
      $schema?: string;
      torrents: components["schemas"]["Torrent"][];
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
      $schema?: string;
      /** Format: int64 */
      addedOn: number;
      /** Format: float */
      availability: number;
      category: string;
      downloadDir: string;
      downloadPath: string;
      /** Format: int64 */
      downloadSpeed: number;
      /** Format: int64 */
      eta: number;
      files: components["schemas"]["FileTree"];
      infoHash: string;
      name: string;
      peers: {
        [key: string]: components["schemas"]["PeerInfo"];
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
    };
    TorrentFileOptions: {
      downloadPriority: string;
      name: string;
      path: string;
    };
    TorrentMeta: {
      /**
       * Format: uri
       * @description A URL to the JSON Schema for this object.
       */
      $schema?: string;
      files: components["schemas"]["FileMeta"][];
      name: string;
      /** Format: int64 */
      totalSize: number;
    };
  };
  responses: never;
  parameters: never;
  requestBodies: never;
  headers: never;
  pathItems: never;
}

export type $defs = Record<string, never>;

export type external = Record<string, never>;

export interface operations {

  /** Get all torrents */
  "get-all-torrents": {
    responses: {
      /** @description OK */
      200: {
        content: {
          "application/json": components["schemas"]["GetTorrentsResBody"];
        };
      };
      /** @description Error */
      default: {
        content: {
          "application/problem+json": components["schemas"]["ErrorModel"];
        };
      };
    };
  };
  /** Post torrent */
  "post-torrent": {
    requestBody: {
      content: {
        "application/json": components["schemas"]["DownloadTorrentReqBody"];
      };
    };
    responses: {
      /** @description OK */
      200: {
        content: {
          "application/json": components["schemas"]["Torrent"];
        };
      };
      /** @description Error */
      default: {
        content: {
          "application/problem+json": components["schemas"]["ErrorModel"];
        };
      };
    };
  };
  /** Post torrent meta */
  "post-torrent-meta": {
    requestBody: {
      content: {
        "application/json": components["schemas"]["GetTorrentMetaReqBody"];
      };
    };
    responses: {
      /** @description OK */
      200: {
        content: {
          "application/json": components["schemas"]["TorrentMeta"];
        };
      };
      /** @description Error */
      default: {
        content: {
          "application/problem+json": components["schemas"]["ErrorModel"];
        };
      };
    };
  };
  /** Get torrent hash */
  "get-torrent-hash": {
    parameters: {
      path: {
        /** @example 2b66980093bc11806fab50cb3cb41835b95a0362 */
        hash: string;
      };
    };
    responses: {
      /** @description OK */
      200: {
        content: {
          "application/json": components["schemas"]["Torrent"];
        };
      };
      /** @description Error */
      default: {
        content: {
          "application/problem+json": components["schemas"]["ErrorModel"];
        };
      };
    };
  };
}