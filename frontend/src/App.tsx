import Layout from "@/Layout";
import HelloWorld from "@/components/hello-world";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import type * as React from "react";
import { Route, Routes } from "react-router";

function App() {
  return (
    <div id="App">
      <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
        <Toaster />
        <Layout>
          <Routes>
            <Route path="/" element={<HelloWorld />} />
            <Route path="/settings" element={<div>Settings Page</div>} />
            <Route path="/search" element={<div>Search Page</div>} />
            <Route path="*" element={<div>404 Not Found</div>} />
          </Routes>
        </Layout>
      </ThemeProvider>
    </div>
  );
}

export default App;
