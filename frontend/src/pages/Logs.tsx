import { Fragment, useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import {
	clearLogs,
	formatTimestamp,
	getLogEntries,
	getLogLevels,
	getLogModules,
	getLogStats,
	getRecentLogs,
	type LogEntry,
	type LogFilter,
	type LogStats,
	parseMetadata,
	searchLogs,
} from "@/services/loggingService";

export default function Logs() {
	const [logs, setLogs] = useState<LogEntry[]>([]);
	const [loading, setLoading] = useState(true);
	const [searchText, setSearchText] = useState("");
	const [selectedModule, setSelectedModule] = useState<string>("");
	const [selectedLevel, setSelectedLevel] = useState<string>("");
	const [stats, setStats] = useState<LogStats | null>(null);
	const [modules, setModules] = useState<string[]>([]);
	const [levels, setLevels] = useState<string[]>([]);
	const [expandedLog, setExpandedLog] = useState<number | null>(null);

	// Load initial data
	useEffect(() => {
		loadInitialData();
	}, []);

	const loadInitialData = useCallback(async () => {
		try {
			setLoading(true);
			const [logsData, statsData, modulesData, levelsData] = await Promise.all([
				getRecentLogs(),
				getLogStats(),
				getLogModules(),
				getLogLevels(),
			]);

			setLogs(logsData);
			setStats(statsData);
			setModules(modulesData);
			setLevels(levelsData);
		} catch (error) {
			console.error("Failed to load log data:", error);
			toast.error("Failed to load log data");
		} finally {
			setLoading(false);
		}
	}, []);

	const handleSearch = async () => {
		if (!searchText.trim()) {
			await loadLogs();
			return;
		}

		try {
			setLoading(true);
			const results = await searchLogs(searchText);
			setLogs(results);
		} catch (error) {
			console.error("Search failed:", error);
			toast.error("Search failed");
		} finally {
			setLoading(false);
		}
	};

	const loadLogs = useCallback(async () => {
		try {
			setLoading(true);
			const filter: LogFilter = {};

			if (selectedModule) filter.module = selectedModule;
			if (selectedLevel) filter.level = selectedLevel;

			const results = await getLogEntries(filter);
			setLogs(results);
		} catch (error) {
			console.error("Failed to load logs:", error);
			toast.error("Failed to load logs");
		} finally {
			setLoading(false);
		}
	}, [selectedModule, selectedLevel]);

	const handleClearLogs = async () => {
		// if (
		// 	!confirm(
		// 		"Are you sure you want to clear all logs? This action cannot be undone.",
		// 	)
		// ) {
		// 	return;
		// }

		try {
			await clearLogs();
			setLogs([]);
			await loadInitialData(); // Reload stats
			toast.success("All logs cleared");
		} catch (error) {
			console.error("Failed to clear logs:", error);
			toast.error("Failed to clear logs");
		}
	};

	const getLogLevelBadgeVariant = (
		level: string,
	): "default" | "secondary" | "destructive" | "outline" => {
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
	};

	const toggleLogExpansion = (logId: number) => {
		setExpandedLog(expandedLog === logId ? null : logId);
	};

	const renderMetadata = (metadata: string) => {
		const parsed = parseMetadata(metadata);
		if (!parsed) return null;

		return (
			<div className="mt-2 p-2 bg-gray-50 dark:bg-gray-800 rounded text-sm">
				<strong>Details:</strong>
				<pre className="mt-1 text-xs overflow-x-auto">
					{JSON.stringify(parsed, null, 2)}
				</pre>
			</div>
		);
	};

	// Apply filters when they change
	useEffect(() => {
		loadLogs();
	}, [loadLogs]);

	if (loading && logs.length === 0) {
		return (
			<div className="container mx-auto p-6">
				<div className="text-center">Loading logs...</div>
			</div>
		);
	}

	return (
		<div className="container mx-auto p-6 space-y-6">
			<div className="flex justify-between items-center">
				<div>
					<h1 className="text-3xl font-bold">Logs</h1>
					<p className="text-muted-foreground">
						Monitor application activity and debug issues
					</p>
				</div>
				<Button variant="destructive" onClick={handleClearLogs}>
					Clear All Logs
				</Button>
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
							<CardTitle className="text-sm font-medium">Error Count</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-2xl font-bold text-red-600">
								{stats.by_level.error || 0}
							</div>
						</CardContent>
					</Card>
				</div>
			)}

			{/* Filters */}
			<Card>
				<CardHeader>
					<CardTitle>Filters & Search</CardTitle>
					<CardDescription>
						Filter logs by module, level, or search text
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
						<div className="space-y-2">
							<Label>Search</Label>
							<div className="flex gap-2">
								<Input
									placeholder="Search logs..."
									value={searchText}
									onChange={(e) => setSearchText(e.target.value)}
									onKeyDown={(e) => e.key === "Enter" && handleSearch()}
								/>
								<Button onClick={handleSearch}>Search</Button>
							</div>
						</div>
						<div className="space-y-2">
							<Label>Module</Label>
							<select
								className="w-full p-2 border rounded"
								value={selectedModule}
								onChange={(e) => setSelectedModule(e.target.value)}
							>
								<option value="">All Modules</option>
								{modules.map((module) => (
									<option key={module} value={module}>
										{module}
									</option>
								))}
							</select>
						</div>
						<div className="space-y-2">
							<Label>Level</Label>
							<select
								className="w-full p-2 border rounded"
								value={selectedLevel}
								onChange={(e) => setSelectedLevel(e.target.value)}
							>
								<option value="">All Levels</option>
								{levels.map((level) => (
									<option key={level} value={level}>
										{level}
									</option>
								))}
							</select>
						</div>
					</div>
				</CardContent>
			</Card>

			{/* Logs Table */}
			<Card>
				<CardHeader>
					<CardTitle>Log Entries</CardTitle>
					<CardDescription>
						{logs.length} entries {loading && "(Loading...)"}
					</CardDescription>
				</CardHeader>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>Timestamp</TableHead>
								<TableHead>Level</TableHead>
								<TableHead>Module</TableHead>
								<TableHead>Message</TableHead>
								<TableHead className="w-[100px]">Details</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{logs.length === 0 ? (
								<TableRow>
									<TableCell colSpan={5} className="text-center py-8">
										No logs found
									</TableCell>
								</TableRow>
							) : (
								logs.map((log) => (
									<Fragment key={log.id}>
										<TableRow className="cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800">
											<TableCell className="font-mono text-xs">
												{formatTimestamp(log.timestamp)}
											</TableCell>
											<TableCell>
												<Badge variant={getLogLevelBadgeVariant(log.level)}>
													{log.level}
												</Badge>
											</TableCell>
											<TableCell>
												<Badge variant="outline">{log.module}</Badge>
											</TableCell>
											<TableCell className="max-w-md truncate">
												{log.message}
											</TableCell>
											<TableCell>
												{log.metadata && (
													<Button
														variant="ghost"
														size="sm"
														onClick={() => toggleLogExpansion(log.id)}
													>
														{expandedLog === log.id ? "Hide" : "Show"}
													</Button>
												)}
											</TableCell>
										</TableRow>
										{expandedLog === log.id && log.metadata && (
											<TableRow>
												<TableCell colSpan={5}>
													{renderMetadata(log.metadata)}
												</TableCell>
											</TableRow>
										)}
									</Fragment>
								))
							)}
						</TableBody>
					</Table>
				</CardContent>
				{logs.length > 0 && (
					<CardFooter>
						<Button variant="outline" onClick={loadLogs} disabled={loading}>
							{loading ? "Loading..." : "Refresh"}
						</Button>
					</CardFooter>
				)}
			</Card>
		</div>
	);
}
