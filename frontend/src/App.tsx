import { GetInitialOpenedCBZ } from "@wails/go/main/App";
import type { main } from "@wails/go/models";
import { Quit, WindowFullscreen, WindowIsFullscreen, WindowUnfullscreen } from "@wails/runtime";
import { EventsOff, EventsOn } from "@wails/runtime/runtime";
import { useEffect } from "preact/hooks";
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
  const { activeRoute, openModal, closeModal, activeModal, navigateTo } = useRouter();
  useOpenCBZ(navigateTo);
  let Route = Home;

  if (activeRoute.name === "comic-view") {
    Route = () => <ComicPage comic={activeRoute.data} />;
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

function useOpenCBZ(navigationFn: ReturnType<typeof useRouter>["navigateTo"]) {
  const eventName = "comic-opened";

  useEffect(() => {
    const handler = (data: main.ComicInfo) => {
      if (!("pageCount" in data) || data?.pageCount <= 0) return;
      navigationFn({ name: "comic-view", data });
    };

    GetInitialOpenedCBZ().then(handler);
    EventsOn(eventName, handler);
    return () => EventsOff(eventName);
  }, []);
}
