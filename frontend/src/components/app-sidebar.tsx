import { NavLink } from "react-router";
import { FileText, Search, Settings } from "lucide-react";

import {
	Sidebar,
	SidebarContent,
	SidebarGroup,
	SidebarGroupContent,
	SidebarGroupLabel,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
} from "@/components/ui/sidebar";

// Menu items.
const items = [
	{
		title: "Settings",
		url: "/settings",
		icon: Settings,
	},
	{
		title: "Search",
		url: "/search",
		icon: Search,
	},
	{
		title: "Logs",
		url: "/logs",
		icon: FileText,
	},
];

export function AppSidebar() {
	return (
		<Sidebar>
			<SidebarContent>
				<SidebarGroup>
					<SidebarGroupLabel>Application</SidebarGroupLabel>
					<SidebarGroupContent>
						<SidebarMenu>
							{items.map((item) => (
								<SidebarMenuItem key={item.title}>
									<NavLink to={item.url}>
										{({ isActive }) => (
											<SidebarMenuButton isActive={isActive}>
												<item.icon />
												<span>{item.title}</span>
											</SidebarMenuButton>
										)}
									</NavLink>
								</SidebarMenuItem>
							))}
						</SidebarMenu>
					</SidebarGroupContent>
				</SidebarGroup>
			</SidebarContent>
		</Sidebar>
	);
}
