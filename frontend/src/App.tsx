import Layout from "@/Layout";
import HelloWorld from "@/components/hello-world";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import type * as React from "react";
import { Route, Routes } from "react-router";
import Settings from "@/pages/Settings";
import LiveSearch from "@/pages/LiveSearch";

function App() {
  return (
    <div id="App">
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
