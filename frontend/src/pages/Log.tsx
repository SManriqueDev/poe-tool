import React, { useEffect, useState } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";

import {
	cleanupOldLogs,
	clearLogs,
	getLogEntries,
	getLogLevels,
	getLogModules,
	getLogStats,
	searchLogs,
} from "../services/loggingService";

interface LogEntry {
	id: number;
	timestamp: string;
	module: string;
	level: string;
	message: string;
	metadata: string;
	created_at: string;
}

interface LogFilter {
	module?: string;
	level?: string;
	search?: string;
	limit?: number;
	offset?: number;
}

interface LogStats {
	total_entries: number;
	today_entries: number;
	by_module: Record<string, number>;
	by_level: Record<string, number>;
}

const levelColors: Record<string, string> = {
	debug: "bg-gray-500",
	info: "bg-blue-500",
	warning: "bg-yellow-500",
	error: "bg-red-500",
	success: "bg-green-500",
};

const levelLabels: Record<string, string> = {
	debug: "Debug",
	info: "Info",
	warning: "Warning",
	error: "Error",
	success: "Success",
};

export default function Log() {
	const [logs, setLogs] = useState<LogEntry[]>([]);
	const [loading, setLoading] = useState(false);
	const [modules, setModules] = useState<string[]>([]);
	const [levels, setLevels] = useState<string[]>([]);
	const [stats, setStats] = useState<LogStats | null>(null);

	// Filters
	const [selectedModule, setSelectedModule] = useState<string>("");
	const [selectedLevel, setSelectedLevel] = useState<string>("");
	const [searchText, setSearchText] = useState("");
	const [limit, setLimit] = useState(100);

	useEffect(() => {
		loadInitialData();
	}, []);

	const loadInitialData = async () => {
		try {
			setLoading(true);

			// Load all initial data in parallel
			const [logsData, modulesData, levelsData, statsData] = await Promise.all([
				getLogEntries({ limit: 100 }),
				getLogModules(),
				getLogLevels(),
				getLogStats(),
			]);

			setLogs(logsData);
			setModules(modulesData);
			setLevels(levelsData);
			setStats(statsData as LogStats);
		} catch (error) {
			console.error("Failed to load log data:", error);
			toast.error("Failed to load log data");
		} finally {
			setLoading(false);
		}
	};

	const handleFilter = async () => {
		try {
			setLoading(true);

			const filter: LogFilter = { limit };

			if (selectedModule) filter.module = selectedModule;
			if (selectedLevel) filter.level = selectedLevel;
			if (searchText) filter.search = searchText;

			const logsData = await getLogEntries(filter);
			setLogs(logsData);
		} catch (error) {
			console.error("Failed to filter logs:", error);
			toast.error("Failed to filter logs");
		} finally {
			setLoading(false);
		}
	};

	const handleClearFilters = () => {
		setSelectedModule("");
		setSelectedLevel("");
		setSearchText("");
		setLimit(100);
		loadInitialData();
	};

	const handleClearLogs = async () => {
		if (
			!confirm(
				"Are you sure you want to clear all logs? This action cannot be undone.",
			)
		) {
			return;
		}

		try {
			await clearLogs();
			toast.success("All logs cleared");
			loadInitialData();
		} catch (error) {
			console.error("Failed to clear logs:", error);
			toast.error("Failed to clear logs");
		}
	};

	const handleCleanup = async () => {
		try {
			await cleanupOldLogs();
			toast.success("Old logs cleaned up");
			loadInitialData();
		} catch (error) {
			console.error("Failed to cleanup logs:", error);
			toast.error("Failed to cleanup logs");
		}
	};

	const formatTimestamp = (timestamp: string) => {
		return new Date(timestamp).toLocaleString();
	};

	const parseMetadata = (metadata: string) => {
		if (!metadata) return null;
		try {
			return JSON.parse(metadata);
		} catch {
			return null;
		}
	};

	const renderMetadata = (metadata: string) => {
		const parsed = parseMetadata(metadata);
		if (!parsed) return null;

		// Handle ItemFoundMetadata
		if (parsed.item_name && parsed.price) {
			return (
				<div className="text-xs text-muted-foreground mt-1">
					<span className="font-medium">{parsed.item_name}</span>
					{parsed.price && (
						<span className="ml-2">
							{parsed.price.amount} {parsed.price.currency}
						</span>
					)}
					{parsed.league && (
						<span className="ml-2 text-blue-600">({parsed.league})</span>
					)}
				</div>
			);
		}

		// Handle APICallMetadata
		if (parsed.url && parsed.status_code) {
			return (
				<div className="text-xs text-muted-foreground mt-1">
					<span>
						{parsed.method} {parsed.url}
					</span>
					<span className="ml-2">Status: {parsed.status_code}</span>
					{parsed.response_time_ms && (
						<span className="ml-2">{parsed.response_time_ms}ms</span>
					)}
				</div>
			);
		}

		// Handle WebSocketMetadata
		if (parsed.search_id && parsed.event_type) {
			return (
				<div className="text-xs text-muted-foreground mt-1">
					<span>WS {parsed.event_type}</span>
					<span className="ml-2">Search: {parsed.search_id}</span>
					{parsed.message_count && (
						<span className="ml-2">{parsed.message_count} messages</span>
					)}
				</div>
			);
		}

		return null;
	};

	return (
		<div className="p-6 space-y-6">
			<div className="flex items-center justify-between">
				<div>
					<h1 className="text-3xl font-bold tracking-tight">Logs</h1>
					<p className="text-muted-foreground">
						View and filter application logs and events
					</p>
				</div>
			</div>

			{/* Stats Cards */}
			{stats && (
				<div className="grid grid-cols-1 md:grid-cols-4 gap-4">
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-sm font-medium">
								Total Entries
							</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-2xl font-bold">{stats.total_entries}</div>
						</CardContent>
					</Card>
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-sm font-medium">Today</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-2xl font-bold">{stats.today_entries}</div>
						</CardContent>
					</Card>
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-sm font-medium">Modules</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-2xl font-bold">
								{Object.keys(stats.by_module).length}
							</div>
						</CardContent>
					</Card>
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-sm font-medium">Levels</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-2xl font-bold">
								{Object.keys(stats.by_level).length}
							</div>
						</CardContent>
					</Card>
				</div>
			)}

			{/* Filters */}
			<Card>
				<CardHeader>
					<CardTitle>Filters</CardTitle>
					<CardDescription>
						Filter logs by module, level, or search text
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="grid grid-cols-1 md:grid-cols-6 gap-4">
						<div>
							<Label htmlFor="module">Module</Label>
							<Select value={selectedModule} onValueChange={setSelectedModule}>
								<SelectTrigger>
									<SelectValue placeholder="All modules" />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="">All modules</SelectItem>
									{modules.map((module) => (
										<SelectItem key={module} value={module}>
											{module}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						<div>
							<Label htmlFor="level">Level</Label>
							<Select value={selectedLevel} onValueChange={setSelectedLevel}>
								<SelectTrigger>
									<SelectValue placeholder="All levels" />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="">All levels</SelectItem>
									{levels.map((level) => (
										<SelectItem key={level} value={level}>
											{levelLabels[level] || level}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						<div>
							<Label htmlFor="search">Search</Label>
							<Input
								id="search"
								placeholder="Search messages..."
								value={searchText}
								onChange={(e) => setSearchText(e.target.value)}
								onKeyDown={(e) => e.key === "Enter" && handleFilter()}
							/>
						</div>

						<div>
							<Label htmlFor="limit">Limit</Label>
							<Select
								value={limit.toString()}
								onValueChange={(v) => setLimit(Number(v))}
							>
								<SelectTrigger>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="50">50</SelectItem>
									<SelectItem value="100">100</SelectItem>
									<SelectItem value="200">200</SelectItem>
									<SelectItem value="500">500</SelectItem>
								</SelectContent>
							</Select>
						</div>

						<div className="flex items-end">
							<Button onClick={handleFilter} disabled={loading}>
								Filter
							</Button>
						</div>

						<div className="flex items-end">
							<Button variant="outline" onClick={handleClearFilters}>
								Clear
							</Button>
						</div>
					</div>

					<Separator className="my-4" />

					<div className="flex space-x-2">
						<Button variant="outline" onClick={handleCleanup}>
							Cleanup Old
						</Button>
						<Button variant="destructive" onClick={handleClearLogs}>
							Clear All
						</Button>
					</div>
				</CardContent>
			</Card>

			{/* Log Entries */}
			<Card>
				<CardHeader>
					<CardTitle>Log Entries</CardTitle>
					<CardDescription>Showing {logs.length} entries</CardDescription>
				</CardHeader>
				<CardContent>
					{loading ? (
						<div className="text-center py-8">Loading...</div>
					) : logs.length === 0 ? (
						<div className="text-center py-8 text-muted-foreground">
							No log entries found
						</div>
					) : (
						<div className="space-y-3">
							{logs.map((log) => (
								<div key={log.id} className="border rounded-lg p-4">
									<div className="flex items-start justify-between">
										<div className="flex-1">
											<div className="flex items-center space-x-2">
												<Badge
													variant="secondary"
													className={`text-white ${levelColors[log.level] || "bg-gray-500"}`}
												>
													{levelLabels[log.level] || log.level}
												</Badge>
												<Badge variant="outline">{log.module}</Badge>
												<span className="text-sm text-muted-foreground">
													{formatTimestamp(log.timestamp)}
												</span>
											</div>
											<div className="mt-2">
												<p className="text-sm font-medium">{log.message}</p>
												{renderMetadata(log.metadata)}
											</div>
										</div>
									</div>
								</div>
							))}
						</div>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
