import { useCallback, useEffect, useState } from "react";
import { Events } from "@wailsio/runtime";
import { RefreshCw, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import SimpleLayout from "@/SimpleLayout";

import {
	cleanupOldLogs,
	getLogEntries,
	type LogEntry,
	type LogFilter,
} from "../services/loggingService";

export default function LiveSearchLogsWindow() {
	const [logs, setLogs] = useState<LogEntry[]>([]);
	const [loading, setLoading] = useState(true);

	const loadLogs = useCallback(async () => {
		try {
			setLoading(true);
			const filter: LogFilter = { module: "livesearch", limit: 1000 };
			const allLogs = await getLogEntries(filter);
			setLogs(allLogs);
		} catch (error) {
			console.error("Error loading LiveSearch logs:", error);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		// Carga inicial
		loadLogs();

		// Escuchar eventos de nuevos logs en tiempo real
		const unsubscribe = Events.On("livesearch:newLog", (ev) => {
			// Los datos vienen como array, necesitamos el primer elemento
			const newLog = ev.data?.[0] as LogEntry;

			if (newLog?.id) {
				setLogs((prevLogs) => {
					// Agregar el nuevo log al principio y limitar a 1000 entradas
					const updatedLogs = [newLog, ...prevLogs].slice(0, 1000);
					return updatedLogs;
				});
			} else {
				console.error(
					"Evento recibido pero el log no tiene el formato esperado:",
					newLog,
				);
			}
		});

		// Cleanup function
		return () => {
			if (unsubscribe) {
				unsubscribe();
			}
		};
	}, [loadLogs]);

	const clearLogs = async () => {
		try {
			await cleanupOldLogs();
			await loadLogs(); // Reload after clearing
		} catch (error) {
			console.error("Error clearing logs:", error);
		}
	};

	const getLevelColor = (level: string) => {
		switch (level.toLowerCase()) {
			case "error":
				return "destructive";
			case "warn":
			case "warning":
				return "secondary";
			case "info":
				return "secondary";
			case "debug":
				return "outline";
			case "success":
				return "default";
			default:
				return "outline";
		}
	};

	return (
		<SimpleLayout>
			<div className="p-6 h-screen flex flex-col">
				<div className="flex justify-between items-center mb-6">
					<div>
						<h1 className="text-2xl font-bold">LiveSearch Logs</h1>
						<p className="text-muted-foreground text-sm">
							Real-time logs for LiveSearch operations and trading activity
						</p>
					</div>
					<div className="flex gap-2">
						<Button
							onClick={loadLogs}
							variant="outline"
							size="sm"
							disabled={loading}
						>
							<RefreshCw
								className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`}
							/>
							Refresh
						</Button>

						<Button onClick={clearLogs} variant="destructive" size="sm">
							<Trash2 className="h-4 w-4 mr-2" />
							Clear
						</Button>
					</div>
				</div>

				<Card className="flex-1 flex flex-col">
					<CardHeader className="pb-3">
						<CardTitle className="flex items-center justify-between text-lg">
							<span>Live Activity</span>
							<div className="flex items-center gap-2">
								<div className="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
								<Badge variant="outline">{logs.length} entries</Badge>
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
									<p className="text-lg mb-2">No LiveSearch logs found</p>
									<p className="text-sm">
										Start a LiveSearch session to see logs here
									</p>
								</div>
							</div>
						) : (
							<div className="flex-1 overflow-y-auto px-6 pb-6">
								<div className="space-y-2">
									{logs.map((log) => (
										<div
											key={`${log.id}-${log.timestamp}`}
											className="flex items-start gap-3 p-4 rounded-lg border bg-card hover:bg-accent/30 transition-colors"
										>
											<Badge
												variant={getLevelColor(log.level)}
												className="w-20 text-xs font-medium justify-center mt-0.5"
											>
												{log.level}
											</Badge>
											<div className="flex-1 min-w-0">
												<div className="flex items-center gap-2 mb-2">
													<span className="font-medium text-sm">
														{log.module}
													</span>
													<span className="text-xs text-muted-foreground">
														{new Date(log.timestamp).toLocaleString()}
													</span>
												</div>
												<p className="text-sm break-words leading-relaxed">
													{log.message}
												</p>
												{log.metadata && log.metadata.trim() !== "" && (
													<div className="mt-3 p-3 bg-muted/50 rounded border text-xs font-mono whitespace-pre-wrap max-h-32 overflow-y-auto">
														{log.metadata}
													</div>
												)}
											</div>
										</div>
									))}
								</div>
							</div>
						)}
					</CardContent>
				</Card>
			</div>
		</SimpleLayout>
	);
}
