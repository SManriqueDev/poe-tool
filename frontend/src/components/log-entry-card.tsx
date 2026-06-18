import { useMemo } from "react";

import { Badge } from "@/components/ui/badge";
import { type LogEntry, getLogLevelBadgeVariant } from "../services/loggingService";

interface LogEntryCardProps {
	log: LogEntry;
}

const MODULE_COLORS: Record<string, string> = {
	livesearch: "text-blue-600 dark:text-blue-400",
	settings: "text-amber-600 dark:text-amber-400",
	api: "text-purple-600 dark:text-purple-400",
	websocket: "text-orange-600 dark:text-orange-400",
	system: "text-slate-600 dark:text-slate-400",
};

function getModuleColor(module: string): string {
	return MODULE_COLORS[module.toLowerCase()] ?? "text-gray-500 dark:text-gray-400";
}

function getRelativeTime(timestamp: string): string {
	const now = Date.now();
	const date = new Date(timestamp).getTime();
	const diffMs = now - date;
	const diffSec = Math.floor(diffMs / 1000);

	if (diffSec < 5) return "just now";
	if (diffSec < 60) return `${diffSec}s ago`;

	const diffMin = Math.floor(diffSec / 60);
	if (diffMin < 60) return `${diffMin}m ago`;

	const diffHr = Math.floor(diffMin / 60);
	if (diffHr < 24) return `${diffHr}h ago`;

	const diffDay = Math.floor(diffHr / 24);
	if (diffDay < 7) return `${diffDay}d ago`;

	return new Date(timestamp).toLocaleDateString();
}

function isUrl(value: string): boolean {
	return value.startsWith("http://") || value.startsWith("https://");
}

interface MetadataBadge {
	key: string;
	value: string;
	isUrl: boolean;
}

function flattenMetadata(obj: Record<string, unknown>, prefix = ""): MetadataBadge[] {
	const result: MetadataBadge[] = [];

	for (const [key, value] of Object.entries(obj)) {
		const fullKey = prefix ? `${prefix}.${key}` : key;

		if (value === null || value === undefined) continue;

		if (typeof value === "object" && !Array.isArray(value)) {
			result.push(...flattenMetadata(value as Record<string, unknown>, fullKey));
		} else if (Array.isArray(value)) {
			result.push({ key: fullKey, value: `[${value.length} items]`, isUrl: false });
		} else {
			const strVal = String(value);
			result.push({
				key: fullKey,
				value: strVal,
				isUrl: isUrl(strVal),
			});
		}
	}

	return result;
}

export function LogEntryCard({ log }: LogEntryCardProps) {
	const metadataBadges = useMemo(() => {
		if (!log.metadata || log.metadata.trim() === "") return null;

		try {
			const parsed = JSON.parse(log.metadata) as Record<string, unknown>;
			if (Object.keys(parsed).length === 0) return null;
			return flattenMetadata(parsed);
		} catch {
			return null;
		}
	}, [log.metadata]);

	const showRawFallback = useMemo(() => {
		if (!log.metadata || log.metadata.trim() === "") return false;
		try {
			JSON.parse(log.metadata);
			return false;
		} catch {
			return true;
		}
	}, [log.metadata]);

	return (
		<div className="flex items-start gap-3 p-4 rounded-lg border bg-card hover:bg-accent/30 transition-colors">
			<Badge
				variant={getLogLevelBadgeVariant(log.level)}
				className="w-20 text-xs font-medium justify-center mt-0.5 shrink-0"
			>
				{log.level}
			</Badge>
			<div className="flex-1 min-w-0">
				<div className="flex items-center gap-2 mb-1">
					<span className={`text-xs font-semibold ${getModuleColor(log.module)}`}>
						{log.module}
					</span>
					<span className="text-xs text-muted-foreground">
						{getRelativeTime(log.timestamp)}
					</span>
				</div>
				<p className="text-sm break-words leading-relaxed">{log.message}</p>
				{metadataBadges && metadataBadges.length > 0 && (
					<div className="flex flex-wrap gap-1.5 mt-2">
						{metadataBadges.map((badge) => (
							<span
								key={badge.key}
								className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs bg-muted/70 text-muted-foreground"
							>
								<span className="font-medium">{badge.key}:</span>
								{badge.isUrl ? (
									<a
										href={badge.value}
										target="_blank"
										rel="noopener noreferrer"
										className="text-primary hover:underline truncate max-w-[200px]"
									>
										{badge.value}
									</a>
								) : (
									<span className="truncate max-w-[200px]">{badge.value}</span>
								)}
							</span>
						))}
					</div>
				)}
				{showRawFallback && (
					<div className="mt-2 p-2 bg-muted/50 rounded border text-xs font-mono whitespace-pre-wrap max-h-24 overflow-y-auto">
						{log.metadata}
					</div>
				)}
			</div>
		</div>
	);
}
