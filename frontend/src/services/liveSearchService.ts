import type { TradeLink as GoTradeLink } from "../../bindings/github.com/SManriqueDev/poe-tool/backend/internal/livesearch/domain/models";
import { Handler } from "../../bindings/github.com/SManriqueDev/poe-tool/backend/internal/livesearch/index.js";

// Define proper TradeLink interface based on the Go model
export interface TradeLink {
	id: number;
	url: string;
	description: string;
	selected: boolean;
	status: string;
}

export const addTradeLink = async (url: string, description: string) => {
	return Handler.AddTradeLink(url, description);
};

export async function deleteTradeLink(id: number): Promise<void> {
	return Handler.DeleteTradeLink(id);
}

export async function updateTradeLink(
	id: number,
	link: TradeLink,
): Promise<void> {
	return Handler.UpdateTradeLink(id, link.url, link.description, link.selected);
}

export async function listTradeLinks(): Promise<GoTradeLink[]> {
	return Handler.ListTradeLinks();
}

export async function startLiveSearch() {
	return Handler.StartLiveSearch();
}

export async function stopLiveSearch() {
	return Handler.StopLiveSearch();
}

export async function setGoToHideout(enabled: boolean): Promise<void> {
	return Handler.SetGoToHideout(enabled);
}

export async function getGoToHideout(): Promise<boolean> {
	return Handler.GetGoToHideout();
}

export async function isLiveSearchRunning(): Promise<boolean> {
	return Handler.IsLiveSearchRunning();
}

// These functions will be available after regenerating Wails bindings
// export async function getHideoutQueueSize(): Promise<number> {
// 	return Handler.GetHideoutQueueSize();
// }

// export async function isHideoutProcessing(): Promise<boolean> {
// 	return Handler.IsHideoutProcessing();
// }

export async function getAllLinkStatuses(): Promise<Record<number, string>> {
	return Handler.GetAllLinkStatuses();
}

export async function openLogsWindow(): Promise<void> {
	return Handler.OpenLogsWindow();
}
