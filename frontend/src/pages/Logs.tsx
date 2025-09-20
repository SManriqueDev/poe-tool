import { Fragment, useCallback, useEffect, useState } from "react";
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
	const [selectedModule, setSelectedModule] = useState<string>("all");
	const [selectedLevel, setSelectedLevel] = useState<string>("all");
	const [stats, setStats] = useState<LogStats | null>(null);
	const [modules, setModules] = useState<string[]>([]);
	const [levels, setLevels] = useState<string[]>([]);
	const [expandedLog, setExpandedLog] = useState<number | null>(null);

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

	// Load initial data
	useEffect(() => {
		loadInitialData();
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [loadInitialData]);

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

			if (selectedModule && selectedModule !== "all")
				filter.module = selectedModule;
			if (selectedLevel && selectedLevel !== "all")
				filter.level = selectedLevel;

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
			<div className="container mx-auto p-4 lg:p-6">
				<div className="text-center text-sm lg:text-base">Loading logs...</div>
			</div>
		);
	}

	return (
		<div className="container mx-auto p-4 lg:p-6 space-y-4 lg:space-y-6">
			<div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
				<div>
					<h1 className="text-2xl lg:text-3xl font-bold">Logs</h1>
					<p className="text-sm lg:text-base text-muted-foreground">
						Monitor application activity and debug issues
					</p>
				</div>
				<Button
					variant="destructive"
					onClick={handleClearLogs}
					size="sm"
					className="self-start sm:self-auto"
				>
					Clear All Logs
				</Button>
			</div>

			{/* Stats Cards */}
			{stats && (
				<div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-xs lg:text-sm font-medium">
								Total Entries
							</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-lg lg:text-2xl font-bold">
								{stats.total_entries}
							</div>
						</CardContent>
					</Card>
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-xs lg:text-sm font-medium">
								Today
							</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-lg lg:text-2xl font-bold">
								{stats.today_entries}
							</div>
						</CardContent>
					</Card>
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-xs lg:text-sm font-medium">
								Modules
							</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-lg lg:text-2xl font-bold">
								{Object.keys(stats.by_module).length}
							</div>
						</CardContent>
					</Card>
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-xs lg:text-sm font-medium">
								Error Count
							</CardTitle>
						</CardHeader>
						<CardContent>
							<div className="text-lg lg:text-2xl font-bold text-red-600">
								{stats.by_level.error || 0}
							</div>
						</CardContent>
					</Card>
				</div>
			)}

			{/* Filters */}
			<Card>
				<CardHeader>
					<CardTitle className="text-lg">Filters & Search</CardTitle>
					<CardDescription className="text-sm">
						Filter logs by module, level, or search text
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					{/* Search Row - Always full width */}
					<div className="space-y-2">
						<Label className="text-sm font-medium">Search</Label>
						<div className="flex gap-2">
							<Input
								placeholder="Search logs..."
								value={searchText}
								onChange={(e) => setSearchText(e.target.value)}
								onKeyDown={(e) => e.key === "Enter" && handleSearch()}
								className="flex-1 min-w-0"
							/>
							<Button
								onClick={handleSearch}
								size="sm"
								className="px-3 lg:px-4 shrink-0"
								disabled={loading}
							>
								Search
							</Button>
						</div>
					</div>

					{/* Filters Row - Responsive grid */}
					<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
						<div className="space-y-2">
							<Label className="text-sm font-medium">Module</Label>
							<Select value={selectedModule} onValueChange={setSelectedModule}>
								<SelectTrigger className="w-full">
									<SelectValue placeholder="All Modules" />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="all">All Modules</SelectItem>
									{modules.map((module) => (
										<SelectItem key={module} value={module}>
											{module.charAt(0).toUpperCase() + module.slice(1)}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						<div className="space-y-2">
							<Label className="text-sm font-medium">Level</Label>
							<Select value={selectedLevel} onValueChange={setSelectedLevel}>
								<SelectTrigger className="w-full">
									<SelectValue placeholder="All Levels" />
								</SelectTrigger>
								<SelectContent>
									<SelectItem value="all">All Levels</SelectItem>
									{levels.map((level) => (
										<SelectItem key={level} value={level}>
											<Badge
												variant={getLogLevelBadgeVariant(level)}
												className="mr-2"
											>
												{level.charAt(0).toUpperCase() + level.slice(1)}
											</Badge>
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</div>

						<div className="space-y-2 sm:col-span-2 lg:col-span-1">
							<Label className="text-sm font-medium">Actions</Label>
							<div className="flex gap-2">
								<Button
									variant="outline"
									onClick={loadLogs}
									disabled={loading}
									size="sm"
									className="flex-1"
								>
									{loading ? "Loading..." : "Refresh"}
								</Button>
								<Button
									variant="ghost"
									onClick={() => {
										setSearchText("");
										setSelectedModule("all");
										setSelectedLevel("all");
									}}
									size="sm"
									className="flex-1"
								>
									Clear
								</Button>
							</div>
						</div>
					</div>
				</CardContent>
			</Card>

			{/* Logs Table */}
			<Card>
				<CardHeader>
					<CardTitle className="text-lg">Log Entries</CardTitle>
					<CardDescription className="text-sm">
						{logs.length} entries {loading && "(Loading...)"}
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="overflow-x-auto">
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
										<TableCell colSpan={5} className="text-center py-8 text-sm">
											No logs found
										</TableCell>
									</TableRow>
								) : (
									logs.map((log) => (
										<Fragment key={log.id}>
											<TableRow className="cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800">
												<TableCell className="font-mono text-xs p-3">
													<div className="truncate">
														{formatTimestamp(log.timestamp)}
													</div>
												</TableCell>
												<TableCell className="p-3">
													<Badge
														variant={getLogLevelBadgeVariant(log.level)}
														className="text-xs"
													>
														{log.level}
													</Badge>
												</TableCell>
												<TableCell className="p-3">
													<Badge variant="outline" className="text-xs">
														{log.module}
													</Badge>
												</TableCell>
												<TableCell className="p-3">
													<div className="max-w-[300px] lg:max-w-[500px] truncate text-sm">
														{log.message}
													</div>
												</TableCell>
												<TableCell className="p-3">
													{log.metadata && (
														<Button
															variant="ghost"
															size="sm"
															onClick={() => toggleLogExpansion(log.id)}
															className="h-8 px-2 text-xs"
														>
															{expandedLog === log.id ? "Hide" : "Show"}
														</Button>
													)}
												</TableCell>
											</TableRow>
											{expandedLog === log.id && log.metadata && (
												<TableRow>
													<TableCell colSpan={5} className="p-3">
														{renderMetadata(log.metadata)}
													</TableCell>
												</TableRow>
											)}
										</Fragment>
									))
								)}
							</TableBody>
						</Table>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
