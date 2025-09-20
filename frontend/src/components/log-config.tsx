import { useEffect, useState } from "react";
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
import { getLogStats, type LogConfig } from "@/services/loggingService";

interface LogConfigComponentProps {
	onConfigChange?: (config: LogConfig) => void;
}

export function LogConfigComponent({
	onConfigChange,
}: LogConfigComponentProps) {
	const [config, setConfig] = useState<LogConfig>({
		enabled: true,
		log_level: "info",
		log_modules: ["livesearch"],
		log_new_items: true,
		log_api_requests: true,
		log_websocket: true,
		retention_days: 30,
		max_entries: 10000,
		real_time_updates: true,
	});

	const [loading, setLoading] = useState(true);

	const logLevels = ["debug", "info", "warning", "error"];
	const logModules = ["livesearch", "settings", "websocket", "api", "system"];

	useEffect(() => {
		loadConfig();
	}, []);

	const loadConfig = async () => {
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
	};

	const handleConfigChange = (newConfig: LogConfig) => {
		setConfig(newConfig);
		onConfigChange?.(newConfig);
	};

	const updateConfig = (updates: Partial<LogConfig>) => {
		const newConfig = { ...config, ...updates };
		handleConfigChange(newConfig);
	};

	const toggleModule = (module: string) => {
		const newModules = config.log_modules.includes(module)
			? config.log_modules.filter((m) => m !== module)
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
						id="logging-enabled"
						checked={config.enabled}
						onCheckedChange={(checked) => updateConfig({ enabled: !!checked })}
					/>
					<Label htmlFor="logging-enabled">Enable logging</Label>
				</div>

				{config.enabled && (
					<>
						{/* Log Level */}
						<div className="space-y-2">
							<Label>Log Level</Label>
							<select
								className="w-full p-2 border rounded"
								value={config.log_level}
								onChange={(e) => updateConfig({ log_level: e.target.value })}
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
									id="log-new-items"
									checked={config.log_new_items}
									onCheckedChange={(checked) =>
										updateConfig({ log_new_items: !!checked })
									}
								/>
								<Label htmlFor="log-new-items">Log new items found</Label>
							</div>

							<div className="flex items-center space-x-2">
								<Checkbox
									id="log-api-requests"
									checked={config.log_api_requests}
									onCheckedChange={(checked) =>
										updateConfig({ log_api_requests: !!checked })
									}
								/>
								<Label htmlFor="log-api-requests">Log API requests</Label>
							</div>

							<div className="flex items-center space-x-2">
								<Checkbox
									id="log-websocket"
									checked={config.log_websocket}
									onCheckedChange={(checked) =>
										updateConfig({ log_websocket: !!checked })
									}
								/>
								<Label htmlFor="log-websocket">Log WebSocket events</Label>
							</div>

							<div className="flex items-center space-x-2">
								<Checkbox
									id="real-time-updates"
									checked={config.real_time_updates}
									onCheckedChange={(checked) =>
										updateConfig({ real_time_updates: !!checked })
									}
								/>
								<Label htmlFor="real-time-updates">Real-time log updates</Label>
							</div>
						</div>

						{/* Retention Settings */}
						<div className="space-y-4">
							<h4 className="font-medium">Log Retention</h4>

							<div className="space-y-2">
								<Label htmlFor="retention-days">Retention Days</Label>
								<Input
									id="retention-days"
									type="number"
									min="1"
									max="365"
									value={config.retention_days}
									onChange={(e) =>
										updateConfig({
											retention_days: parseInt(e.target.value) || 30,
										})
									}
								/>
								<p className="text-sm text-muted-foreground">
									Automatically delete logs older than this many days
								</p>
							</div>

							<div className="space-y-2">
								<Label htmlFor="max-entries">Maximum Entries</Label>
								<Input
									id="max-entries"
									type="number"
									min="100"
									max="100000"
									value={config.max_entries}
									onChange={(e) =>
										updateConfig({
											max_entries: parseInt(e.target.value) || 10000,
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
