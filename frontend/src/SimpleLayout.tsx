import type { ReactNode } from "react";

import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";

interface SimpleLayoutProps {
	children: ReactNode;
}

export default function SimpleLayout({ children }: SimpleLayoutProps) {
	return (
		<ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
			<Toaster />
			<main className="min-h-screen bg-background">
				{children}
			</main>
		</ThemeProvider>
	);
}