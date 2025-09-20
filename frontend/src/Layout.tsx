import type * as React from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { ModeToggle } from "@/components/mode-toggle";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";

export default function Layout({ children }: { children: React.ReactNode }) {
	return (
		<SidebarProvider>
			<AppSidebar />
			<SidebarTrigger />
			<main className="relative flex min-h-svh overflow-auto w-full flex-col pr-6">
				{children}
			</main>
			<div className="fixed bottom-4 right-4">
				<ModeToggle />
			</div>
		</SidebarProvider>
	);
}
