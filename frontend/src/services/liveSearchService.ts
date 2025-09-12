import {
  AddTradeLink,
  UpdateTradeLinks,
  ListTradeLinks,
} from "../../wailsjs/go/livesearch/Handler";
import { livesearch } from "../../wailsjs/go/models";
import TradeLink = livesearch.TradeLink;

export async function addTradeLink(url: string, description: string): Promise<void> {
  return AddTradeLink(url, description);
}

export async function updateTradeLinks(links: TradeLink[]): Promise<void> {
  return UpdateTradeLinks(links);
}

export async function listTradeLinks(): Promise<TradeLink[]> {
  return ListTradeLinks();
}
