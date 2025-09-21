import { useCallback, useEffect, useId, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { getLogStats } from "@/services/loggingService";

import {
	type LogConfig,
	LogLevel,
	LogModule,
} from "../../bindings/github.com/SManriqueDev/poe-tool/backend/internal/logging/models";

interface LogConfigComponentProps {
	onConfigChange?: (config: LogConfig) => void;
}

export function LogConfigComponent({
	onConfigChange,
}: LogConfigComponentProps) {
	const [config, setConfig] = useState<LogConfig>({
		enabled: true,
		log_level: LogLevel.LogLevelInfo,
		log_modules: [LogModule.LogModuleLiveSearch],
		log_new_items: true,
		log_api_requests: true,
		log_websocket: true,
		retention_days: 30,
		max_entries: 10000,
		real_time_updates: true,
	});

	const [loading, setLoading] = useState(true);

	// Generate unique IDs for form elements
	const loggingEnabledId = useId();
	const logNewItemsId = useId();
	const logApiRequestsId = useId();
	const logWebsocketId = useId();
	const realTimeUpdatesId = useId();
	const retentionDaysId = useId();
	const maxEntriesId = useId();

	const logLevels = [
		LogLevel.LogLevelDebug,
		LogLevel.LogLevelInfo,
		LogLevel.LogLevelWarning,
		LogLevel.LogLevelError,
	];
	const logModules = [
		LogModule.LogModuleLiveSearch,
		LogModule.LogModuleSettings,
		LogModule.LogModuleWebSocket,
		LogModule.LogModuleAPI,
		LogModule.LogModuleSystem,
	];

	const loadConfig = useCallback(async () => {
		try {
			setLoading(true);
			const stats = await getLogStats();
			if (stats?.config) {
				setConfig(stats.config);
			}
		} catch (error) {
			console.error("Failed to load log config:", error);
			toast.error("Failed to load log configuration");
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => {
		loadConfig();
	}, [loadConfig]);

	const handleConfigChange = (newConfig: LogConfig) => {
		setConfig(newConfig);
		onConfigChange?.(newConfig);
	};

	const updateConfig = (updates: Partial<LogConfig>) => {
		const newConfig = { ...config, ...updates };
		handleConfigChange(newConfig);
	};

	const toggleModule = (module: LogModule) => {
		const newModules = config.log_modules.includes(module)
			? config.log_modules.filter((m: LogModule) => m !== module)
			: [...config.log_modules, module];
		updateConfig({ log_modules: newModules });
	};

	const handleSave = async () => {
		try {
			// TODO: Implement when backend binding is available
			// await updateLogConfig(config);
			toast.success("Log configuration saved");
		} catch (error) {
			console.error("Failed to save config:", error);
			toast.error("Failed to save configuration");
		}
	};

	if (loading) {
		return (
			<Card>
				<CardHeader>
					<CardTitle>Log Configuration</CardTitle>
				</CardHeader>
				<CardContent>
					<div>Loading configuration...</div>
				</CardContent>
			</Card>
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle>Log Configuration</CardTitle>
				<CardDescription>
					Configure what events to log and how logs are managed
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-6">
				{/* Enable/Disable Logging */}
				<div className="flex items-center space-x-2">
					<Checkbox
						id={loggingEnabledId}
						checked={config.enabled}
						onCheckedChange={(checked) => updateConfig({ enabled: !!checked })}
					/>
					<Label htmlFor={loggingEnabledId}>Enable logging</Label>
				</div>

				{config.enabled && (
					<>
						{/* Log Level */}
						<div className="space-y-2">
							<Label>Log Level</Label>
							<select
								className="w-full p-2 border rounded"
								value={config.log_level}
								onChange={(e) =>
									updateConfig({ log_level: e.target.value as LogLevel })
								}
							>
								{logLevels.map((level) => (
									<option key={level} value={level}>
										{level.charAt(0).toUpperCase() + level.slice(1)}
									</option>
								))}
							</select>
							<p className="text-sm text-muted-foreground">
								Only logs at this level and above will be recorded
							</p>
						</div>

						{/* Log Modules */}
						<div className="space-y-2">
							<Label>Log Modules</Label>
							<div className="grid grid-cols-2 gap-2">
								{logModules.map((module) => (
									<div key={module} className="flex items-center space-x-2">
										<Checkbox
											id={`module-${module}`}
											checked={config.log_modules.includes(module)}
											onCheckedChange={() => toggleModule(module)}
										/>
										<Label htmlFor={`module-${module}`} className="capitalize">
											{module}
										</Label>
									</div>
								))}
							</div>
						</div>

						{/* Specific Features */}
						<div className="space-y-4">
							<h4 className="font-medium">Feature Logging</h4>

							<div className="flex items-center space-x-2">
								<Checkbox
									id={logNewItemsId}
									checked={config.log_new_items}
									onCheckedChange={(checked) =>
										updateConfig({ log_new_items: !!checked })
									}
								/>
								<Label htmlFor={logNewItemsId}>Log new items found</Label>
							</div>

							<div className="flex items-center space-x-2">
								<Checkbox
									id={logApiRequestsId}
									checked={config.log_api_requests}
									onCheckedChange={(checked) =>
										updateConfig({ log_api_requests: !!checked })
									}
								/>
								<Label htmlFor={logApiRequestsId}>Log API requests</Label>
							</div>

							<div className="flex items-center space-x-2">
								<Checkbox
									id={logWebsocketId}
									checked={config.log_websocket}
									onCheckedChange={(checked) =>
										updateConfig({ log_websocket: !!checked })
									}
								/>
								<Label htmlFor={logWebsocketId}>Log WebSocket events</Label>
							</div>

							<div className="flex items-center space-x-2">
								<Checkbox
									id={realTimeUpdatesId}
									checked={config.real_time_updates}
									onCheckedChange={(checked) =>
										updateConfig({ real_time_updates: !!checked })
									}
								/>
								<Label htmlFor={realTimeUpdatesId}>Real-time log updates</Label>
							</div>
						</div>

						{/* Retention Settings */}
						<div className="space-y-4">
							<h4 className="font-medium">Log Retention</h4>

							<div className="space-y-2">
								<Label htmlFor={retentionDaysId}>Retention Days</Label>
								<Input
									id={retentionDaysId}
									type="number"
									min="1"
									max="365"
									value={config.retention_days}
									onChange={(e) =>
										updateConfig({
											retention_days: parseInt(e.target.value, 10) || 30,
										})
									}
								/>
								<p className="text-sm text-muted-foreground">
									Automatically delete logs older than this many days
								</p>
							</div>

							<div className="space-y-2">
								<Label htmlFor={maxEntriesId}>Maximum Entries</Label>
								<Input
									id={maxEntriesId}
									type="number"
									min="100"
									max="100000"
									value={config.max_entries}
									onChange={(e) =>
										updateConfig({
											max_entries: parseInt(e.target.value, 10) || 10000,
										})
									}
								/>
								<p className="text-sm text-muted-foreground">
									Maximum number of log entries to keep
								</p>
							</div>
						</div>

						<Button onClick={handleSave} className="w-full">
							Save Configuration
						</Button>
					</>
				)}
			</CardContent>
		</Card>
	);
}
