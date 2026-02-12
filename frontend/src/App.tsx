import { ComicPage } from "./components/ComicPage";
import { ComicProvider } from "./components/ComicProvider";
import { Home } from "./components/Home";
import { RouterProvider, useRouter } from "./components/RouterProvider";

export function App() {
  return (
    <RouterProvider>
      <ComicProvider>
        <RouterController />
      </ComicProvider>
    </RouterProvider>
  );
}

function RouterController() {
  const { activeRoute } = useRouter();
  let Route = Home;

  const comicPrefix = "comic-id: ";
  if (activeRoute.startsWith(comicPrefix)) {
    const comicId = activeRoute.slice(comicPrefix.length);
    Route = () => <ComicPage comicId={comicId} />;
  }

  return <Route />;
}
