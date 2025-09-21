import { Handler } from "../../bindings/github.com/SManriqueDev/poe-tool/backend/internal/settings/index.js";
import type { Config } from "../../bindings/github.com/SManriqueDev/poe-tool/backend/internal/settings/models.js";

export async function loadConfig(): Promise<Config | null> {
	return await Handler.LoadConfig();
}

export async function saveConfig(cfg: Config): Promise<void> {
	await Handler.SaveConfig(cfg);
}
