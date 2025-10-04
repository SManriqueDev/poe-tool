import type React from "react";
import { useEffect, useId, useState } from "react";
import { Events } from "@wailsio/runtime";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { getColumns } from "@/live-search/columns";
import { DataTable } from "@/live-search/data-table";
import {
	addTradeLink,
	isLiveSearchRunning as checkLiveSearchStatus,
	deleteTradeLink,
	getAllLinkStatuses,
	getGoToHideout,
	listTradeLinks,
	openLogsWindow,
	setGoToHideout,
	startLiveSearch,
	stopLiveSearch,
	type TradeLink,
	updateTradeLink,
} from "@/services/liveSearchService";
import {
	formatTimestamp,
	getLogEntries,
	type LogEntry,
	parseMetadata,
} from "@/services/loggingService";

// Type definitions for the application
interface ItemResult {
	id: string;
	item: Record<string, unknown>;
	listing: Record<string, unknown>;
}

interface NewItemsFoundEventData {
	searchID: string;
	items: ItemResult[];
	count: number;
}

interface WailsEvent<T = unknown> {
	data: T[];
}

// interface NewItemsEvent {
// 	searchID: string;
// 	items: {
// 		id: string;
// 		item: Record<string, unknown>;
// 		listing: Record<string, unknown>;
// 	}[];
// 	count: number;
// }

export default function LiveSearch() {
	const checkboxId = useId();
	const [url, setUrl] = useState("");
	const [description, setDescription] = useState("");
	const [links, setLinks] = useState<TradeLink[]>([]);
	const [editIdx, setEditIdx] = useState<number | null>(null);
	const [editUrl, setEditUrl] = useState("");
	const [editDescription, setEditDescription] = useState("");
	const [isLiveSearchRunning, setIsLiveSearchRunning] = useState(false);
	const [goToHideoutEnabled, setGoToHideoutEnabled] = useState(false);

	// Log viewer state
	const [showLogs, setShowLogs] = useState(false);
	const [logs, setLogs] = useState<LogEntry[]>([]);
	const [logsLoading, setLogsLoading] = useState(false);

	// Load LiveSearch-specific logs
	const loadLiveSearchLogs = async () => {
		try {
			setLogsLoading(true);
			const liveSearchLogs = await getLogEntries({
				module: "livesearch",
				limit: 100,
			});
			setLogs(liveSearchLogs);
		} catch (error) {
			console.error("Failed to load LiveSearch logs:", error);
			toast.error("Failed to load logs");
		} finally {
			setLogsLoading(false);
		}
	};

	// Toggle log viewer
	const toggleLogs = async () => {
		if (!showLogs) {
			await loadLiveSearchLogs();
		}
		setShowLogs(!showLogs);
	};

	// Open logs in new window
	const handleOpenLogsWindow = async () => {
		try {
			await openLogsWindow();
		} catch (error) {
			console.error("Failed to open logs window:", error);
			toast.error("Failed to open logs window");
		}
	};

	// Get badge color for log levels (consistent with LiveSearchLogsWindow)
	const getLevelColor = (level: string) => {
		switch (level.toLowerCase()) {
			case "error":
				return "destructive";
			case "warn":
			case "warning":
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

	useEffect(() => {
		// Load trade links and their current statuses
		Promise.all([listTradeLinks(), getAllLinkStatuses()])
			.then(([links, statuses]) => {
				// Apply statuses to links
				const linksWithStatus = links.map((link) => {
					const finalStatus = statuses[link.id] || link.status || "idle";
					return {
						...link,
						status: finalStatus,
					};
				});

				setLinks(linksWithStatus);
			})
			.catch((error) => {
				console.error("Failed to load trade links or statuses:", error);
				// Fallback to just loading links
				listTradeLinks().then((links) => {
					setLinks(links);
				});
			});

		// Check current live search status from backend
		checkLiveSearchStatus()
			.then((running) => {
				setIsLiveSearchRunning(running);
			})
			.catch((error) => {
				console.error("Failed to check live search status:", error);
			});

		// Load go to hideout setting
		getGoToHideout()
			.then((enabled) => {
				setGoToHideoutEnabled(enabled);
			})
			.catch((error) => {
				console.error("Failed to load go to hideout setting:", error);
			});

		const offStatusChanged = Events.On(
			"linkStatusChanged",
			(ev: WailsEvent<TradeLink>) => {
				const link = ev.data[0] || ev.data;
				setLinks((prev) => {
					const updated = prev.map((l) => {
						if (l.id === link.id) {
							// Solo actualizar el status, no sobrescribir todo el objeto
							return { ...l, status: link.status };
						}
						return l;
					});
					return updated;
				});
			},
		);

		const offNewItems = Events.On(
			"livesearch:newItemsFound",
			(ev: WailsEvent<NewItemsFoundEventData>) => {
				// In Wails v3, the actual data is in ev.data[0]
				const data = ev.data[0] || ev.data;

				// Show toast notification
				toast(`Found ${data.count} new item${data.count > 1 ? "s" : ""}!`, {
					description: `Search ID: ${data.searchID}`,
					action: {
						label: "View",
						onClick: () => {
							// Here you could open a modal or navigate to item details
							console.log("View items clicked", data.items);
						},
					},
				});
			},
		);

		return () => {
			offStatusChanged();
			offNewItems();
		};
	}, []);

	const handleAdd = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!url) return;
		await addTradeLink(url, description);
		setUrl("");
		setDescription("");
		// Recargar links pero preservar statuses actuales
		const newLinks = await listTradeLinks();
		const statuses = await getAllLinkStatuses();
		const linksWithStatus = newLinks.map((link) => ({
			...link,
			status: statuses[link.id] || link.status || "idle",
		}));
		setLinks(linksWithStatus);
		toast("Link added!");
	};

	const handleDelete = async (idx: number) => {
		const updated = links.filter((_, i) => i !== idx);
		setLinks(updated);
		await deleteTradeLink(links[idx].id);
		toast("Link deleted!");
	};

	const handleEdit = (idx: number) => {
		setEditIdx(idx);
		setEditUrl(links[idx].url);
		setEditDescription(links[idx].description || "");
	};

	const handleSaveEdit = async (idx: number) => {
		const link = links[idx];
		await updateTradeLink(link.id, {
			...link,
			url: editUrl,
			description: editDescription,
		} as TradeLink);
		// Recargar links pero preservar statuses actuales
		const newLinks = await listTradeLinks();
		const statuses = await getAllLinkStatuses();
		const linksWithStatus = newLinks.map((link) => ({
			...link,
			status: statuses[link.id] || link.status || "idle",
		}));
		setLinks(linksWithStatus);
		setEditIdx(null);
		toast("Link updated!");
	};

	const handleCancelEdit = () => {
		setEditIdx(null);
	};

	const handleStart = async () => {
		setIsLiveSearchRunning(true);
		toast("Starting live search for selected links...");
		await startLiveSearch();

		// Los status updates vendrán por eventos
		// Verificar errores de autenticación se hará a través de eventos de status
	};

	const handleStop = async () => {
		await stopLiveSearch();
		setIsLiveSearchRunning(false);
		toast("Live search stopped.");
	};

	const handleSelect = async (idx: number, selected: boolean) => {
		const link = links[idx];
		await updateTradeLink(link.id, { ...link, selected } as TradeLink);
		const updatedLinks = await listTradeLinks();
		setLinks(updatedLinks);
	};

	const handleGoToHideoutChange = async (checked: boolean) => {
		try {
			await setGoToHideout(checked);
			setGoToHideoutEnabled(checked);
			toast.success(`Auto-visit hideout ${checked ? "enabled" : "disabled"}`);
		} catch (error) {
			console.error("Failed to update go to hideout setting:", error);
			toast.error("Failed to update setting");
		}
	};

	return (
		<div className="max-w-3xl w-full mx-auto mt-12 space-y-4">
			<Card>
				<CardHeader>
					<div className="flex justify-between items-center">
						<CardTitle>Live Search</CardTitle>
						<div className="flex gap-2">
							<Button
								variant="outline"
								size="sm"
								onClick={toggleLogs}
								disabled={logsLoading}
							>
								{logsLoading
									? "Loading..."
									: showLogs
										? "Hide Logs"
										: "View Logs"}
							</Button>
							<Button
								variant="outline"
								size="sm"
								onClick={handleOpenLogsWindow}
							>
								Open Logs Window
							</Button>
						</div>
					</div>
				</CardHeader>
				<CardContent>
					<form className="flex gap-4 mb-6" onSubmit={handleAdd}>
						<Input
							placeholder="Trade link URL"
							value={url}
							onChange={(e) => setUrl(e.target.value)}
							required
						/>
						<Input
							placeholder="Description (optional)"
							value={description}
							onChange={(e) => setDescription(e.target.value)}
						/>

						<Button type="submit">Add</Button>
					</form>

					<div className="flex items-center space-x-2 mb-4 p-3 bg-muted/30 rounded-lg">
						<Checkbox
							id={checkboxId}
							checked={goToHideoutEnabled}
							onCheckedChange={handleGoToHideoutChange}
						/>
						<Label
							htmlFor={checkboxId}
							className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
						>
							Automatically visit seller's hideout when trade opportunity is
							found
						</Label>
					</div>

					<DataTable
						columns={getColumns({
							editIdx,
							editUrl,
							editDescription,
							setEditUrl,
							setEditDescription,
							handleEdit,
							handleSaveEdit,
							handleCancelEdit,
							handleDelete,
							handleSelect,
							data: links,
						})}
						data={links}
					/>
				</CardContent>
				<CardFooter>
					{isLiveSearchRunning ? (
						<Button
							className="w-full"
							variant="destructive"
							onClick={handleStop}
						>
							Stop Live Search
						</Button>
					) : (
						<Button className="w-full" onClick={handleStart}>
							Start Live Search
						</Button>
					)}
				</CardFooter>
			</Card>

			{/* LiveSearch Logs Viewer */}
			{showLogs && (
				<Card>
					<CardHeader>
						<div className="flex justify-between items-center">
							<CardTitle className="text-lg">LiveSearch Logs</CardTitle>
							<Badge variant="secondary">{logs.length} entries</Badge>
						</div>
					</CardHeader>
					<CardContent>
						{logsLoading ? (
							<div className="text-center py-8">Loading logs...</div>
						) : logs.length === 0 ? (
							<div className="text-center py-8 text-muted-foreground">
								No LiveSearch logs found
							</div>
						) : (
							<div className="space-y-4">
								<div className="text-sm text-muted-foreground mb-4">
									Showing recent LiveSearch activity and events
								</div>
								<div className="max-h-96 overflow-y-auto">
									<Table>
										<TableHeader>
											<TableRow>
												<TableHead className="w-[140px]">Time</TableHead>
												<TableHead className="w-[80px]">Level</TableHead>
												<TableHead>Message</TableHead>
											</TableRow>
										</TableHeader>
										<TableBody>
											{logs.map((log) => (
												<TableRow key={log.id}>
													<TableCell className="font-mono text-xs">
														{formatTimestamp(log.timestamp)}
													</TableCell>
													<TableCell>
														<Badge
															variant={getLevelColor(log.level)}
															className="w-20 text-xs font-medium justify-center"
														>
															{log.level}
														</Badge>
													</TableCell>
													<TableCell>
														<div className="space-y-1">
															<div className="text-sm">{log.message}</div>
															{log.metadata && parseMetadata(log.metadata) && (
																<div className="text-xs text-muted-foreground">
																	{(() => {
																		const metadata = parseMetadata(
																			log.metadata,
																		);
																		if (metadata?.item_name) {
																			return `Item: ${metadata.item_name}`;
																		}
																		if (metadata?.search_id) {
																			return `Search ID: ${metadata.search_id}`;
																		}
																		if (metadata?.url) {
																			return `URL: ${metadata.url}`;
																		}
																		return null;
																	})()}
																</div>
															)}
														</div>
													</TableCell>
												</TableRow>
											))}
										</TableBody>
									</Table>
								</div>
								<Separator />
								<div className="flex justify-between items-center text-sm text-muted-foreground">
									<span>Last updated: {new Date().toLocaleTimeString()}</span>
									<Button
										variant="ghost"
										size="sm"
										onClick={loadLiveSearchLogs}
									>
										Refresh
									</Button>
								</div>
							</div>
						)}
					</CardContent>
				</Card>
			)}
		</div>
	);
}
