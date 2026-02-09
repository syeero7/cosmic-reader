import { ComicProvider } from "./components/ComicProvider";
import { Home } from "./components/Home";

export function App() {
  return (
    <ComicProvider>
      <Home />
    </ComicProvider>
  );
}
