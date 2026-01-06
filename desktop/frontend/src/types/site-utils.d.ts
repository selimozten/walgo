/**
 * Site Utility Types
 *
 * These types correspond to the Go backend types in:
 * - pkg/api/api_site_utils.go
 * - desktop/app.go
 */

/**
 * Result of site path validation
 */
export interface ValidateSitePathResult {
  /** Whether the path is valid */
  valid: boolean;
  /** Whether this is a Hugo site */
  isHugoSite: boolean;
  /** Whether walgo.yaml exists */
  hasWalgoConfig: boolean;
  /** Path to content directory */
  contentDir: string;
  /** Path to publish directory */
  publishDir: string;
  /** List of validation issues */
  issues: string[];
  /** Error message if validation failed */
  error?: string;
}

/**
 * Statistics about a Hugo site
 */
export interface GetSiteStatsResult {
  /** Whether stats retrieval was successful */
  success: boolean;
  /** Total number of content files */
  totalFiles: number;
  /** Map of content type name to file count */
  contentTypes: Record<string, number>;
  /** Total size of content in bytes */
  totalSize: number;
  /** Whether the site has been built */
  isBuilt: boolean;
  /** Size of published output in bytes */
  publishedSize: number;
  /** Error message if stats retrieval failed */
  error?: string;
}

/**
 * Result of site initialization
 */
export interface InitializeSiteResult {
  /** Whether initialization was successful */
  success: boolean;
  /** Path to the initialized site */
  sitePath: string;
  /** Success/info message */
  message: string;
  /** Error message if initialization failed */
  error?: string;
}

/**
 * Summary item for a content type
 */
export interface ContentTypeSummaryItem {
  /** Name of the content type */
  name: string;
  /** Number of files in this type */
  fileCount: number;
  /** Whether this is the default type */
  isDefault: boolean;
  /** Up to 5 most recent filenames */
  recentFiles: string[];
  /** Sample post titles (if available) */
  sampleTitles: string[];
}

/**
 * Summary of all content types in a site
 */
export interface GetContentTypeSummaryResult {
  /** Whether summary retrieval was successful */
  success: boolean;
  /** Array of content type summaries */
  types: ContentTypeSummaryItem[];
  /** Total files across all types */
  totalFiles: number;
  /** Name of the default content type */
  defaultType: string;
  /** Error message if summary retrieval failed */
  error?: string;
}

/**
 * Extended window interface with site utility methods
 */
declare global {
  interface Window {
    go: {
      main: {
        App: {
          /**
           * Validates if a path is a valid Hugo/Walgo site
           * @param sitePath Path to validate
           * @returns Validation result with issues
           */
          ValidateSitePath(sitePath: string): Promise<ValidateSitePathResult>;

          /**
           * Gets statistics about a Hugo site
           * @param sitePath Path to the Hugo site
           * @returns Site statistics including file counts and sizes
           */
          GetSiteStats(sitePath: string): Promise<GetSiteStatsResult>;

          /**
           * Initializes walgo.yaml in an existing Hugo site
           * @param sitePath Path to the Hugo site
           * @returns Initialization result
           */
          InitializeWalgoConfig(sitePath: string): Promise<InitializeSiteResult>;

          /**
           * Gets a detailed summary of content types with recent files
           * @param sitePath Path to the Hugo site
           * @returns Content type summary with recent files
           */
          GetContentTypeSummary(
            sitePath: string
          ): Promise<GetContentTypeSummaryResult>;

          // Other existing methods...
          [key: string]: any;
        };
      };
    };
  }
}

export {};
