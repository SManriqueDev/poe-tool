import type * as React from "react";
import { FileText } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Button } from "@/components/ui/button";
import { ModeToggle } from "@/components/mode-toggle";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { openLogsWindow } from "./services/loggingService";

export default function Layout({ children }: { children: React.ReactNode }) {
	return (
		<SidebarProvider>
			<AppSidebar />
			<SidebarTrigger />
			<main className="relative flex min-h-svh overflow-auto w-full flex-col pr-6">
				{children}
			</main>
			<div className="fixed bottom-4 right-4 flex items-center gap-2">
				<Button
					variant="outline"
					size="icon"
					onClick={openLogsWindow}
					title="Open logs window"
				>
					<FileText className="h-4 w-4" />
				</Button>
				<ModeToggle />
			</div>
		</SidebarProvider>
	);
}
