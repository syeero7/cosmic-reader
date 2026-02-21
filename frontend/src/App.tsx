import { Quit, WindowFullscreen, WindowIsFullscreen, WindowUnfullscreen } from "@wails/runtime";
import { ComicPage } from "./components/ComicPage";
import { ComicProvider } from "./components/ComicProvider";
import { Home } from "./components/Home";
import { KbdShortcutModal } from "./components/KbdShortcutModal";
import { RouterProvider, useRouter } from "./components/RouterProvider";
import { ShortcutProvider, useShortcut } from "./components/ShortcutProvider";

export function App() {
  return (
    <RouterProvider>
      <ComicProvider>
        <ShortcutProvider>
          <RouterController />
          <KbdShortcutModal />
        </ShortcutProvider>
      </ComicProvider>
    </RouterProvider>
  );
}

function RouterController() {
  const { activeRoute, openModal, closeModal, activeModal } = useRouter();
  let Route = Home;

  const comicPrefix = "comic-id: ";
  if (activeRoute.startsWith(comicPrefix)) {
    const comicId = activeRoute.slice(comicPrefix.length);
    Route = () => <ComicPage comicId={comicId} />;
  }

  useShortcut("Control+Shift+f", toggleFullscreen);
  useShortcut("Control+q", Quit);

  useShortcut("Alt+k", () => {
    if (activeModal === "kbd-shortcuts") {
      closeModal();
      return;
    }

    openModal("kbd-shortcuts");
  });

  return <Route />;
}

async function toggleFullscreen() {
  if (await WindowIsFullscreen()) {
    WindowUnfullscreen();
    return;
  }

  WindowFullscreen();
}
