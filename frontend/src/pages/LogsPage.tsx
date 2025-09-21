import { useCallback, useEffect, useState } from "react";
import { RefreshCw, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import {
	cleanupOldLogs,
	getLogEntries,
	type LogEntry,
	type LogFilter,
} from "../services/loggingService";

export default function LogsPage() {
	const [logs, setLogs] = useState<LogEntry[]>([]);
	const [loading, setLoading] = useState(true);
	const [source, setSource] = useState<string>("");

	const loadLogs = useCallback(async () => {
		try {
			setLoading(true);
			const filter: LogFilter =
				source === "livesearch"
					? { module: "livesearch", limit: 500 }
					: { limit: 500 };
			const allLogs = await getLogEntries(filter);
			setLogs(allLogs);
		} catch (error) {
			console.error("Error loading logs:", error);
		} finally {
			setLoading(false);
		}
	}, [source]);

	useEffect(() => {
		// Obtener el parámetro source de la URL
		const urlParams = new URLSearchParams(window.location.search);
		const sourceParam = urlParams.get("source") || "all";
		setSource(sourceParam);
	}, []);

	useEffect(() => {
		if (source) {
			loadLogs();
		}
	}, [source, loadLogs]);

	const clearLogs = async () => {
		try {
			await cleanupOldLogs();
			setLogs([]);
		} catch (error) {
			console.error("Error clearing logs:", error);
		}
	};

	const getLevelColor = (level: string) => {
		switch (level.toLowerCase()) {
			case "error":
				return "destructive";
			case "warn":
				return "secondary";
			case "info":
				return "default";
			case "debug":
				return "outline";
			default:
				return "default";
		}
	};

	const filteredLogs =
		source === "livesearch"
			? logs.filter((log) => log.module.includes("livesearch"))
			: logs;

	return (
		<div className="container mx-auto p-6">
			<div className="flex justify-between items-center mb-6">
				<div>
					<h1 className="text-3xl font-bold">
						{source === "livesearch" ? "LiveSearch Logs" : "Application Logs"}
					</h1>
					<p className="text-muted-foreground">
						{source === "livesearch"
							? "Logs específicos de LiveSearch y operaciones de trading"
							: "Todos los logs de la aplicación"}
					</p>
				</div>
				<div className="flex gap-2">
					<Button onClick={loadLogs} variant="outline" size="sm">
						<RefreshCw className="h-4 w-4 mr-2" />
						Refresh
					</Button>
					<Button onClick={clearLogs} variant="destructive" size="sm">
						<Trash2 className="h-4 w-4 mr-2" />
						Clear Logs
					</Button>
				</div>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="flex items-center justify-between">
						<span>Log Entries</span>
						<Badge variant="outline">{filteredLogs.length} entries</Badge>
					</CardTitle>
				</CardHeader>
				<CardContent>
					{loading ? (
						<div className="text-center py-8">
							<RefreshCw className="h-8 w-8 animate-spin mx-auto mb-2" />
							<p>Loading logs...</p>
						</div>
					) : filteredLogs.length === 0 ? (
						<div className="text-center py-8 text-muted-foreground">
							<p>No logs found</p>
							{source === "livesearch" && (
								<p className="text-sm mt-2">
									Start a LiveSearch session to see logs here
								</p>
							)}
						</div>
					) : (
						<div className="space-y-2 max-h-[600px] overflow-y-auto">
							{filteredLogs.map((log) => (
								<div
									key={`${log.id}-${log.timestamp}`}
									className="flex items-start gap-3 p-3 rounded-lg border bg-card hover:bg-accent/50 transition-colors"
								>
									<Badge variant={getLevelColor(log.level)} className="mt-0.5">
										{log.level.toUpperCase()}
									</Badge>
									<div className="flex-1 min-w-0">
										<div className="flex items-center gap-2 mb-1">
											<span className="font-medium text-sm">{log.module}</span>
											<span className="text-xs text-muted-foreground">
												{new Date(log.timestamp).toLocaleString()}
											</span>
										</div>
										<p className="text-sm break-words">{log.message}</p>
										{log.metadata && log.metadata.trim() !== "" && (
											<div className="mt-2 p-2 bg-muted rounded text-xs font-mono whitespace-pre-wrap">
												{log.metadata}
											</div>
										)}
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
