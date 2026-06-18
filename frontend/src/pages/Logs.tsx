import { useCallback, useEffect, useRef, useState } from "react";
import { Events } from "@wailsio/runtime";
import { FileText, RefreshCw, Search, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import SimpleLayout from "@/SimpleLayout";

import { LogEntryCard } from "../components/log-entry-card";
import {
	clearLogs as clearAllLogs,
	getLogEntries,
	getLogEntriesCount,
	getLogLevels,
	getLogModules,
	type LogEntry,
	type LogFilter,
} from "../services/loggingService";

const PAGE_SIZE = 50;

export default function Logs() {
	const [logs, setLogs] = useState<LogEntry[]>([]);
	const [loading, setLoading] = useState(true);
	const [modules, setModules] = useState<string[]>([]);
	const [levels, setLevels] = useState<string[]>([]);
	const [page, setPage] = useState(1);
	const [totalCount, setTotalCount] = useState(0);
	const [selectedModule, setSelectedModule] = useState("");
	const [selectedLevel, setSelectedLevel] = useState("");
	const [searchText, setSearchText] = useState("");
	const unsubscribeRef = useRef<(() => void) | undefined>(undefined);

	const totalPages = Math.max(1, Math.ceil(totalCount / PAGE_SIZE));

	const applyFilter = useCallback(
		async (module: string, level: string, search: string, pageNum: number) => {
			try {
				setLoading(true);
				const filter: LogFilter = {
					limit: PAGE_SIZE,
					offset: (pageNum - 1) * PAGE_SIZE,
				};
				if (module) filter.module = module;
				if (level) filter.level = level;
				if (search) filter.search = search;

				const [entries, count] = await Promise.all([
					getLogEntries(filter),
					getLogEntriesCount(filter),
				]);
				setLogs(entries);
				setTotalCount(count);
			} catch (error) {
				console.error("Failed to load logs:", error);
			} finally {
				setLoading(false);
			}
		},
		[],
	);

	const loadInitialData = useCallback(async () => {
		try {
			const [mods, lvls] = await Promise.all([
				getLogModules(),
				getLogLevels(),
			]);
			setModules(mods);
			setLevels(lvls);
		} catch (error) {
			console.error("Failed to load filter options:", error);
		}
	}, []);

	useEffect(() => {
		loadInitialData();
		applyFilter("", "", "", 1);

		const unsubscribe = Events.On("logs:newEntry", (ev) => {
			const newLog = ev.data?.[0] as LogEntry;
			if (newLog?.id) {
				setLogs((prevLogs) => {
					const updatedLogs = [newLog, ...prevLogs].slice(0, 1000);
					return updatedLogs;
				});
				setTotalCount((prev) => prev + 1);
			}
		});

		unsubscribeRef.current = unsubscribe;

		return () => {
			if (unsubscribeRef.current) {
				unsubscribeRef.current();
			}
		};
	}, [loadInitialData, applyFilter]);

	const handleModuleChange = (value: string) => {
		setSelectedModule(value);
		setPage(1);
		applyFilter(value, selectedLevel, searchText, 1);
	};

	const handleLevelChange = (value: string) => {
		setSelectedLevel(value);
		setPage(1);
		applyFilter(selectedModule, value, searchText, 1);
	};

	const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const value = e.target.value;
		setSearchText(value);
		setPage(1);
		applyFilter(selectedModule, selectedLevel, value, 1);
	};

	const handlePageChange = (newPage: number) => {
		if (newPage < 1 || newPage > totalPages) return;
		setPage(newPage);
		applyFilter(selectedModule, selectedLevel, searchText, newPage);
	};

	const handleRefresh = () => {
		applyFilter(selectedModule, selectedLevel, searchText, page);
	};

	const handleClear = async () => {
		try {
			await clearAllLogs();
			setLogs([]);
			setTotalCount(0);
			setPage(1);
		} catch (error) {
			console.error("Failed to clear logs:", error);
		}
	};

	const renderPageNumbers = () => {
		const pages: React.ReactNode[] = [];
		const maxVisible = 5;
		let start = Math.max(1, page - Math.floor(maxVisible / 2));
		const end = Math.min(totalPages, start + maxVisible - 1);
		if (end - start + 1 < maxVisible) {
			start = Math.max(1, end - maxVisible + 1);
		}

		if (page > 1) {
			pages.push(
				<Button key="prev" variant="outline" size="sm" onClick={() => handlePageChange(page - 1)}>
					Prev
				</Button>,
			);
		}

		for (let i = start; i <= end; i++) {
			pages.push(
				<Button
					key={i}
					variant={i === page ? "default" : "outline"}
					size="sm"
					onClick={() => handlePageChange(i)}
				>
					{i}
				</Button>,
			);
		}

		if (page < totalPages) {
			pages.push(
				<Button key="next" variant="outline" size="sm" onClick={() => handlePageChange(page + 1)}>
					Next
				</Button>,
			);
		}

		return pages;
	};

	return (
		<SimpleLayout>
			<div className="p-6 h-screen flex flex-col">
				<div className="flex justify-between items-center mb-4">
					<div>
						<h1 className="text-2xl font-bold">Logs</h1>
						<p className="text-muted-foreground text-sm">
							Real-time logs for all application modules
						</p>
					</div>
					<div className="flex gap-2">
						<Button
							onClick={handleRefresh}
							variant="outline"
							size="sm"
							disabled={loading}
						>
							<RefreshCw
								className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`}
							/>
							Refresh
						</Button>
						<Button onClick={handleClear} variant="destructive" size="sm">
							<Trash2 className="h-4 w-4 mr-2" />
							Clear
						</Button>
					</div>
				</div>

				<div className="flex gap-2 mb-4">
					<div className="w-44">
						<Select value={selectedModule} onValueChange={handleModuleChange}>
							<SelectTrigger>
								<SelectValue placeholder="All modules" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">All modules</SelectItem>
								{modules.map((m) => (
									<SelectItem key={m} value={m}>
										{m}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
					<div className="w-36">
						<Select value={selectedLevel} onValueChange={handleLevelChange}>
							<SelectTrigger>
								<SelectValue placeholder="All levels" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">All levels</SelectItem>
								{levels.map((l) => (
									<SelectItem key={l} value={l}>
										{l}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
					<div className="flex-1 relative">
						<Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
						<Input
							placeholder="Search logs..."
							value={searchText}
							onChange={handleSearchChange}
							className="pl-9"
						/>
					</div>
				</div>

				<Card className="flex-1 flex flex-col">
					<CardHeader className="pb-3">
						<CardTitle className="flex items-center justify-between text-lg">
							<span>Live Activity</span>
							<div className="flex items-center gap-2">
								<div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
								<Badge variant="outline">
									{totalCount} {totalCount === 1 ? "entry" : "entries"}
								</Badge>
							</div>
						</CardTitle>
					</CardHeader>
					<CardContent className="flex-1 flex flex-col p-0">
						{loading && logs.length === 0 ? (
							<div className="flex-1 flex items-center justify-center">
								<div className="text-center">
									<RefreshCw className="h-8 w-8 animate-spin mx-auto mb-2 text-muted-foreground" />
									<p className="text-muted-foreground">Loading logs...</p>
								</div>
							</div>
						) : logs.length === 0 ? (
							<div className="flex-1 flex items-center justify-center">
								<div className="text-center text-muted-foreground">
									<FileText className="h-12 w-12 mx-auto mb-3 opacity-50" />
									<p className="text-lg mb-1">No logs found</p>
									<p className="text-sm">Adjust filters or start a tool to see activity</p>
								</div>
							</div>
						) : (
							<div className="flex-1 overflow-y-auto px-6 pb-6">
								<div className="space-y-2">
									{logs.map((log) => (
										<LogEntryCard key={`${log.id}-${log.timestamp}`} log={log} />
									))}
								</div>
							</div>
						)}
					</CardContent>
				</Card>

				{totalPages > 1 && (
					<div className="flex items-center justify-center gap-1 mt-4">
						{renderPageNumbers()}
						<span className="text-sm text-muted-foreground ml-2">
							Page {page} of {totalPages}
						</span>
					</div>
				)}
			</div>
		</SimpleLayout>
	);
}
