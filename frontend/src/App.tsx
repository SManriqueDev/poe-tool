import { useId } from "react";
import { Route, Routes } from "react-router";

import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import Layout from "@/Layout";
import LiveSearch from "@/pages/LiveSearch";
import LiveSearchLogsWindow from "@/pages/LiveSearchLogsWindow";
import Logs from "@/pages/Logs";
import Settings from "@/pages/Settings";

function App() {
	return (
		<div id={useId()}>
			<Routes>
				{/* Rutas sin layout para ventanas independientes */}
				<Route path="/livesearch-logs" element={<LiveSearchLogsWindow />} />

				{/* Rutas principales con layout */}
				<Route
					path="/*"
					element={
						<ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
							<Toaster />
							<Layout>
								<Routes>
									<Route path="/" element={<Settings />} />
									<Route path="/settings" element={<Settings />} />
									<Route path="/search" element={<LiveSearch />} />
									<Route path="/logs" element={<Logs />} />
									<Route path="*" element={<div>404 Not Found</div>} />
								</Routes>
							</Layout>
						</ThemeProvider>
					}
				/>
			</Routes>
		</div>
	);
}

export default App;
