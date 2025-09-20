import { AppSidebar } from "@/components/app-sidebar";
import { ModeToggle } from "@/components/mode-toggle";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import type * as React from "react";

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarTrigger />
      <main className="flex min-h-screen w-full flex-col pr-6">{children}</main>
      <div className="fixed bottom-4 right-4">
        <ModeToggle />
      </div>
    </SidebarProvider>
  );
}
