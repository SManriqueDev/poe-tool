import {
	AddTradeLink,
	DeleteTradeLink,
	GetGoToHideout,
	ListTradeLinks,
	SetGoToHideout,
	StartLiveSearch,
	StopLiveSearch,
	UpdateTradeLink,
} from "@wails/go/livesearch/Handler";
import { livesearch } from "@wails/go/models";

import TradeLink = livesearch.TradeLink;

export async function addTradeLink(
	url: string,
	description: string,
): Promise<void> {
	return AddTradeLink(url, description);
}

export async function deleteTradeLink(id: number): Promise<void> {
	return DeleteTradeLink(id);
}

export async function updateTradeLink(
	id: number,
	link: TradeLink,
): Promise<void> {
	return UpdateTradeLink(id, link.url, link.description, link.selected);
}

export async function listTradeLinks(): Promise<TradeLink[]> {
	return ListTradeLinks();
}

export async function startLiveSearch() {
	return StartLiveSearch();
}

export async function stopLiveSearch() {
	return StopLiveSearch();
}

export async function setGoToHideout(enabled: boolean): Promise<void> {
	return SetGoToHideout(enabled);
}

export async function getGoToHideout(): Promise<boolean> {
	return GetGoToHideout();
}
