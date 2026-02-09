import { GetComicInfo } from "@wails/go/main/App";
import { createContext } from "preact";
import { type PropsWithChildren, useContext, useEffect, useReducer } from "preact/compat";

type ComicCtx = {
  comics: Readonly<Awaited<ReturnType<typeof GetComicInfo>>>;
  dispatch: (action: ReducerAction) => void;
};

const ComicContext = createContext<ComicCtx | null>(null);

export function useComics() {
  const ctx = useContext(ComicContext);
  if (!ctx) throw new Error("useComics hook must be used within a child of ComicProvider");
  return ctx;
}

export function ComicProvider({ children }: PropsWithChildren) {
  const [comics, dispatch] = useReducer(reducer, []);
  useEffect(() => {
    (async () => {
      const res = await GetComicInfo();
      dispatch({ type: "add_comics", payload: res });
    })();
  }, []);
  return <ComicContext.Provider value={{ comics, dispatch }}>{children}</ComicContext.Provider>;
}

type ReducerAction = {
  type: "add_comics" | "delete_comic";
  payload: ComicCtx["comics"] | ReturnType<typeof crypto.randomUUID>;
};

function reducer(state: ComicCtx["comics"], action: ReducerAction) {
  switch (action.type) {
    case "add_comics": {
      if (typeof action.payload === "string") {
        throw new Error(`invalid payload: ${action.payload}`);
      }

      return [...state, ...action.payload];
    }
    case "delete_comic": {
      return state.filter((c) => c.id !== action.payload);
    }
  }
}
