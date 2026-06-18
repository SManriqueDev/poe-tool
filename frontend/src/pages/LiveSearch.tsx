import type React from "react";
import { useEffect, useId, useReducer, useState } from "react";
import { Events } from "@wailsio/runtime";
import { toast } from "sonner";

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
import { getColumns } from "@/live-search/columns";
import { DataTable } from "@/live-search/data-table";
import {
	addTradeLink,
	isLiveSearchRunning as checkLiveSearchStatus,
	deleteTradeLink,
	getAllLinkStatuses,
	getGoToHideout,
	listTradeLinks,
	setGoToHideout,
	startLiveSearch,
	stopLiveSearch,
	type TradeLink,
	updateTradeLink,
} from "@/services/liveSearchService";

// Type definitions for the application
interface EnrichedItemResult {
	id: string;
	item: Record<string, unknown>;
	listing: Record<string, unknown>;
	itemName?: string;
	itemTypeLine?: string;
	priceAmount?: number;
	priceCurrency?: string;
	priceType?: string;
}

interface NewItemsFoundEventData {
	searchId: string;
	items: EnrichedItemResult[];
}

interface WailsEvent<T = unknown> {
	data: T[];
}

// Reducer para manejar actualizaciones de estado de links de manera thread-safe
type LinkAction =
	| { type: "UPDATE_STATUS"; linkID: number; status: string }
	| { type: "SET_LINKS"; links: TradeLink[] }
	| { type: "UPDATE_LINK"; link: TradeLink };

const linkReducer = (state: TradeLink[], action: LinkAction): TradeLink[] => {
	switch (action.type) {
		case "UPDATE_STATUS":
			return state.map((l) =>
				l.id === action.linkID ? { ...l, status: action.status } : l,
			);
		case "SET_LINKS":
			return action.links;
		case "UPDATE_LINK":
			return state.map((l) => (l.id === action.link.id ? action.link : l));
		default:
			return state;
	}
};

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
	const [links, dispatch] = useReducer(linkReducer, []);
	const [editIdx, setEditIdx] = useState<number | null>(null);
	const [editUrl, setEditUrl] = useState("");
	const [editDescription, setEditDescription] = useState("");
	const [isLiveSearchRunning, setIsLiveSearchRunning] = useState(false);
	const [goToHideoutEnabled, setGoToHideoutEnabled] = useState(false);

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

				dispatch({ type: "SET_LINKS", links: linksWithStatus });
			})
			.catch((error) => {
				console.error("Failed to load trade links or statuses:", error);
				// Fallback to just loading links
				listTradeLinks().then((links) => {
					dispatch({ type: "SET_LINKS", links });
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
			(ev: WailsEvent<{ linkID: number; status: string }>) => {
				const data = ev.data[0] || ev.data;

				dispatch({
					type: "UPDATE_STATUS",
					linkID: data.linkID,
					status: data.status,
				});
			},
		);

		const offNewItems = Events.On(
			"livesearch:new-items",
			(ev: WailsEvent<NewItemsFoundEventData>) => {
				const data = ev.data[0] || ev.data;
				const first = data.items?.[0];

				// Determine title: item name, fall back to typeLine, then generic
				const title = first?.itemName || first?.itemTypeLine || "New item";

				// Build description with price if available
				let description = `Search: ${data.searchId}`;
				if (first?.priceAmount != null && first?.priceCurrency) {
					description = `${first.priceAmount} ${first.priceCurrency} — ${description}`;
				}

				toast(title, {
					description,
					action: {
						label: "View",
						onClick: () => {
							console.log("View items clicked", data.items);
						},
					},
				});
			},
		);

		const offLiveSearchStarted = Events.On("livesearch:started", () => {
			console.log("Live search started event received");
			setIsLiveSearchRunning(true);
			toast("Live search started!");
		});

		const offLiveSearchStopped = Events.On("livesearch:stopped", () => {
			setIsLiveSearchRunning(false);
			toast("Live search stopped!");
		});

		return () => {
			offStatusChanged();
			offNewItems();
			offLiveSearchStarted();
			offLiveSearchStopped();
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
		dispatch({ type: "SET_LINKS", links: linksWithStatus });
		toast("Link added!");
	};

	const handleDelete = async (idx: number) => {
		const updated = links.filter((_, i) => i !== idx);
		dispatch({ type: "SET_LINKS", links: updated });
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
		dispatch({ type: "SET_LINKS", links: linksWithStatus });
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
		dispatch({ type: "SET_LINKS", links: updatedLinks });
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
					<CardTitle>Live Search</CardTitle>
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
		</div>
	);
}
