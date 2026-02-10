import { ComicProvider } from "./components/ComicProvider";
import { Home } from "./components/Home";
import { RouterProvider } from "./components/RouterProvider";

export function App() {
  return (
    <RouterProvider>
      <ComicProvider>
        <Home />
      </ComicProvider>
    </RouterProvider>
  );
}
