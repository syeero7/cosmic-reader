import type { main } from "@wails/go/models";
import { createContext } from "preact";
import type { PropsWithChildren } from "preact/compat";
import { useContext, useState } from "preact/hooks";

type Modals = "kbd-shortcuts" | "app-error";
type Route = { name: "home" } | { name: "comic-view"; data: main.ComicInfo };
type RouterCtx = {
  navigateTo: (to: Route) => void;
  openModal: (modal: Modals) => void;
  closeModal: () => void;
  activeRoute: Route;
  activeModal?: Modals;
};

const RouterContext = createContext<RouterCtx | null>(null);

export function useRouter() {
  const ctx = useContext(RouterContext);
  if (!ctx) throw new Error("useComics hook must be used within a child of ComicProvider");
  return ctx;
}

export function RouterProvider({ children }: PropsWithChildren) {
  const [activeModal, setActiveModal] = useState<Modals | undefined>(undefined);
  const [activeRoute, setActiveRoute] = useState<Route>({ name: "home" });

  const value: RouterCtx = {
    navigateTo: setActiveRoute,
    openModal: (v: Modals) => setActiveModal(v),
    closeModal: () => setActiveModal(undefined),
    activeRoute,
    activeModal,
  };

  return <RouterContext.Provider value={value}>{children}</RouterContext.Provider>;
}
