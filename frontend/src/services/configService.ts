import { LoadConfig, SaveConfig } from "../../wailsjs/go/settings/Handler";
import { settings } from "../../wailsjs/go/models";
import Config = settings.Config;

export async function loadConfig(): Promise<Config> {
  return await LoadConfig();
}

export async function saveConfig(cfg: Config): Promise<void> {
  await SaveConfig(cfg);
}
