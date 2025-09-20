import type React from "react";
import { useEffect, useId, useState } from "react";
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
	deleteTradeLink,
	getGoToHideout,
	listTradeLinks,
	setGoToHideout,
	startLiveSearch,
	stopLiveSearch,
	updateTradeLink,
} from "@/services/liveSearchService";
import { livesearch } from "~wails/go/models";
import { EventsOn } from "~wails/runtime";

import TradeLink = livesearch.TradeLink;

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

	useEffect(() => {
		listTradeLinks().then((links) => {
			console.log("Fetched trade links", links);
			setLinks(links);
		});

		// Load go to hideout setting
		getGoToHideout()
			.then((enabled) => {
				setGoToHideoutEnabled(enabled);
			})
			.catch((error) => {
				console.error("Failed to load go to hideout setting:", error);
			});

		const off = EventsOn("linkStatusChanged", (link: TradeLink) => {
			console.log("Received linkStatusChanged event", link);
			setLinks((prev) =>
				prev.map((l) => (l.id === link.id ? { ...l, ...link } : l)),
			);
		});
		return () => {
			off();
		};
	}, []);

	const handleAdd = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!url) return;
		await addTradeLink(url, description);
		setUrl("");
		setDescription("");
		setLinks(await listTradeLinks());
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
		setLinks(await listTradeLinks());
		setEditIdx(null);
		toast("Link updated!");
	};

	const handleCancelEdit = () => {
		setEditIdx(null);
	};

	const handleStart = async () => {
		setIsLiveSearchRunning(true);
		toast("Starting live search for selected links...");
		const updatedLinks = await startLiveSearch();
		console.log("Live search started, updated links:", updatedLinks);
		setLinks(updatedLinks);
		if (updatedLinks.some((link) => link.status === "auth_error")) {
			toast.error(
				"Your POESESSID is invalid or expired. Please update it in settings.",
			);
			setIsLiveSearchRunning(false);
		}
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
		<Card className="max-w-3xl w-full mx-auto mt-12">
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
						Automatically visit seller's hideout when trade opportunity is found
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
					<Button className="w-full" variant="destructive" onClick={handleStop}>
						Stop Live Search
					</Button>
				) : (
					<Button className="w-full" onClick={handleStart}>
						Start Live Search
					</Button>
				)}
			</CardFooter>
		</Card>
	);
}
