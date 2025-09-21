import { Handler } from "../../bindings/github.com/SManriqueDev/poe-tool/backend/internal/settings/index.js";

export async function loadConfig(): Promise<any> {
	return await Handler.LoadConfig();
}

export async function saveConfig(cfg: any): Promise<void> {
	await Handler.SaveConfig(cfg);
}
