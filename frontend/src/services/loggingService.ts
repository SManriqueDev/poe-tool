import {
	CleanupOldLogs,
	ClearLogs,
	GetLogConfig,
	GetLogEntries,
	GetLogEntriesCount,
	GetLogLevels,
	GetLogModules,
	GetLogStats,
	GetRecentLogs,
	SearchLogs,
	UpdateLogConfig,
} from "../../wailsjs/go/logging/Handler";
import { logging } from "../../wailsjs/go/models";

// Re-export types from Wails models for consistency
export type LogEntry = {
	id: number;
	timestamp: string;
	module: string;
	level: string;
	message: string;
	metadata: string;
	created_at: string;
};

export type LogFilter = {
	module?: string;
	level?: string;
	start_time?: string;
	end_time?: string;
	search?: string;
	limit?: number;
	offset?: number;
};

export type LogConfig = {
	enabled: boolean;
	log_level: string;
	log_modules: string[];
	log_new_items: boolean;
	log_api_requests: boolean;
	log_websocket: boolean;
	retention_days: number;
	max_entries: number;
	real_time_updates: boolean;
};

export interface LogStats {
	total_entries: number;
	today_entries: number;
	by_module: Record<string, number>;
	by_level: Record<string, number>;
	config: LogConfig;
}

// Helper function to convert Wails LogEntry to our LogEntry type
function convertLogEntry(entry: logging.LogEntry): LogEntry {
	return {
		id: entry.id,
		timestamp: entry.timestamp?.toString() || "",
		module: entry.module,
		level: entry.level,
		message: entry.message,
		metadata: entry.metadata,
		created_at: entry.created_at?.toString() || "",
	};
}

// Helper function to convert our LogFilter to Wails LogFilter
function convertLogFilter(filter: LogFilter): logging.LogFilter {
	const wailsFilter = new logging.LogFilter();
	wailsFilter.module = filter.module;
	wailsFilter.level = filter.level;
	wailsFilter.search = filter.search;
	wailsFilter.limit = filter.limit;
	wailsFilter.offset = filter.offset;
	// Note: start_time and end_time would need proper Date conversion if used
	return wailsFilter;
}

// Logging service functions
export async function getLogEntries(
	filter: LogFilter = {},
): Promise<LogEntry[]> {
	try {
		const wailsFilter = convertLogFilter(filter);
		const entries = await GetLogEntries(wailsFilter);
		return entries.map(convertLogEntry);
	} catch (error) {
		console.error("Failed to get log entries:", error);
		return [];
	}
}

export async function getLogEntriesCount(
	filter: LogFilter = {},
): Promise<number> {
	try {
		const wailsFilter = convertLogFilter(filter);
		return await GetLogEntriesCount(wailsFilter);
	} catch (error) {
		console.error("Failed to get log entries count:", error);
		return 0;
	}
}

export async function getLogModules(): Promise<string[]> {
	try {
		return await GetLogModules();
	} catch (error) {
		console.error("Failed to get log modules:", error);
		return ["livesearch", "settings", "websocket", "api", "system"];
	}
}

export async function getLogLevels(): Promise<string[]> {
	try {
		return await GetLogLevels();
	} catch (error) {
		console.error("Failed to get log levels:", error);
		return ["debug", "info", "warning", "error", "success"];
	}
}

export async function getRecentLogs(): Promise<LogEntry[]> {
	try {
		const entries = await GetRecentLogs();
		return entries.map(convertLogEntry);
	} catch (error) {
		console.error("Failed to get recent logs:", error);
		return [];
	}
}

export async function searchLogs(
	searchText: string,
	limit = 50,
): Promise<LogEntry[]> {
	try {
		const entries = await SearchLogs(searchText, limit);
		return entries.map(convertLogEntry);
	} catch (error) {
		console.error("Failed to search logs:", error);
		return [];
	}
}

export async function clearLogs(): Promise<void> {
	try {
		await ClearLogs();
	} catch (error) {
		console.error("Failed to clear logs:", error);
		throw error;
	}
}

export async function getLogStats(): Promise<LogStats> {
	try {
		const stats = await GetLogStats();
		// GetLogStats returns Record<string, any>, so we need to map it to our LogStats interface
		return {
			total_entries: stats.total_entries || 0,
			today_entries: stats.today_entries || 0,
			by_module: stats.by_module || {},
			by_level: stats.by_level || {},
			config: stats.config || {
				enabled: true,
				log_level: "info",
				log_modules: ["livesearch"],
				log_new_items: true,
				log_api_requests: true,
				log_websocket: true,
				retention_days: 30,
				max_entries: 10000,
				real_time_updates: true,
			},
		};
	} catch (error) {
		console.error("Failed to get log stats:", error);
		return {
			total_entries: 0,
			today_entries: 0,
			by_module: {},
			by_level: {},
			config: {
				enabled: true,
				log_level: "info",
				log_modules: ["livesearch"],
				log_new_items: true,
				log_api_requests: true,
				log_websocket: true,
				retention_days: 30,
				max_entries: 10000,
				real_time_updates: true,
			},
		};
	}
}

export async function getLogConfig(): Promise<LogConfig> {
	try {
		const config = await GetLogConfig();
		return {
			enabled: config.enabled,
			log_level: config.log_level,
			log_modules: config.log_modules,
			log_new_items: config.log_new_items,
			log_api_requests: config.log_api_requests,
			log_websocket: config.log_websocket,
			retention_days: config.retention_days,
			max_entries: config.max_entries,
			real_time_updates: config.real_time_updates,
		};
	} catch (error) {
		console.error("Failed to get log config:", error);
		return {
			enabled: true,
			log_level: "info",
			log_modules: ["livesearch"],
			log_new_items: true,
			log_api_requests: true,
			log_websocket: true,
			retention_days: 30,
			max_entries: 10000,
			real_time_updates: true,
		};
	}
}

export async function updateLogConfig(config: LogConfig): Promise<void> {
	try {
		const wailsConfig = new logging.LogConfig();
		wailsConfig.enabled = config.enabled;
		wailsConfig.log_level = config.log_level;
		wailsConfig.log_modules = config.log_modules;
		wailsConfig.log_new_items = config.log_new_items;
		wailsConfig.log_api_requests = config.log_api_requests;
		wailsConfig.log_websocket = config.log_websocket;
		wailsConfig.retention_days = config.retention_days;
		wailsConfig.max_entries = config.max_entries;
		wailsConfig.real_time_updates = config.real_time_updates;

		await UpdateLogConfig(wailsConfig);
	} catch (error) {
		console.error("Failed to update log config:", error);
		throw error;
	}
}

export async function cleanupOldLogs(): Promise<void> {
	try {
		await CleanupOldLogs();
	} catch (error) {
		console.error("Failed to cleanup old logs:", error);
		throw error;
	}
}

// Helper functions
export function formatTimestamp(timestamp: string): string {
	try {
		const date = new Date(timestamp);
		return date.toLocaleString();
	} catch {
		return timestamp;
	}
}

export function getLogLevelColor(level: string): string {
	switch (level.toLowerCase()) {
		case "debug":
			return "text-gray-500";
		case "info":
			return "text-blue-500";
		case "warning":
			return "text-yellow-500";
		case "error":
			return "text-red-500";
		case "success":
			return "text-green-500";
		default:
			return "text-gray-500";
	}
}

export function getLogLevelBadgeVariant(
	level: string,
): "default" | "secondary" | "destructive" | "outline" {
	switch (level.toLowerCase()) {
		case "error":
			return "destructive";
		case "warning":
			return "outline";
		case "success":
			return "default";
		default:
			return "secondary";
	}
}

export function parseMetadata(
	metadata: string,
): Record<string, unknown> | null {
	if (!metadata || metadata.trim() === "") {
		return null;
	}
	try {
		return JSON.parse(metadata) as Record<string, unknown>;
	} catch {
		return null;
	}
}
