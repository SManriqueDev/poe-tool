import Layout from "@/Layout";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import LiveSearch from "@/pages/LiveSearch";
import Settings from "@/pages/Settings";
import { useId } from "react";
import { Route, Routes } from "react-router";

function App() {
  return (
    <div id={useId()}>
      <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
        <Toaster />
        <Layout>
          <Routes>
            <Route path="/" element={<Settings />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="/search" element={<LiveSearch />} />
            <Route path="*" element={<div>404 Not Found</div>} />
          </Routes>
        </Layout>
      </ThemeProvider>
    </div>
  );
}

export default App;
