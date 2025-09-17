import {
    AddTradeLink,
    UpdateTradeLink,
    ListTradeLinks,
    StopLiveSearch,
    StartLiveSearch,
    SetGoToHideout,
    GetGoToHideout
} from "../../wailsjs/go/livesearch/Handler";
import {livesearch} from "../../wailsjs/go/models";
import TradeLink = livesearch.TradeLink;

export async function addTradeLink(url: string, description: string): Promise<void> {
    return AddTradeLink(url, description);
}

// export async function updateTradeLinks(links: TradeLink[]): Promise<void> {
//   return UpdateTradeLinks(links);
// }

export async function updateTradeLink(id: number, link: TradeLink): Promise<void> {
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

export async function setGoToHideout(value: boolean) {
    return SetGoToHideout(value);
}

export async function getGoToHideout(): Promise<boolean> {
    return GetGoToHideout();
}