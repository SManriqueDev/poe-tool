import { settings } from "@wails/go/models";
import { LoadConfig, SaveConfig } from "@wails/go/settings/Handler";

import Config = settings.Config;

export async function loadConfig(): Promise<Config> {
	return await LoadConfig();
}

export async function saveConfig(cfg: Config): Promise<void> {
	await SaveConfig(cfg);
}
