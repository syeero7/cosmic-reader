import { GetComicList } from "@wails/go/main/App";
import { createContext } from "preact";
import { type PropsWithChildren, useContext, useEffect, useReducer } from "preact/compat";

export type ComicInfo = { id: string; title: string; pageCount: number };

type ComicCtx = {
  comics: Record<string, string>;
  dispatch: (action: ReducerAction) => void;
};

const ComicContext = createContext<ComicCtx | null>(null);

export function useComics() {
  const ctx = useContext(ComicContext);
  if (!ctx) throw new Error("useComics hook must be used within a child of ComicProvider");
  return ctx;
}

export function ComicProvider({ children }: PropsWithChildren) {
  const [comics, dispatch] = useReducer(reducer, {});
  useEffect(() => {
    (async () => {
      const res = await GetComicList();
      dispatch({ type: "add_comics", payload: res });
    })();
  }, []);
  return <ComicContext.Provider value={{ comics, dispatch }}>{children}</ComicContext.Provider>;
}

type ReducerAction = {
  type: "add_comics" | "delete_comic";
  payload: ComicCtx["comics"] | string;
};

function reducer(state: ComicCtx["comics"], action: ReducerAction) {
  switch (action.type) {
    case "add_comics": {
      if (typeof action.payload === "string") {
        throw new Error(`invalid payload: ${action.payload}`);
      }

      return { ...state, ...action.payload };
    }
    case "delete_comic": {
      if (typeof action.payload !== "string") {
        throw new Error(`invalid payload: ${action.payload}`);
      }

      const { [action.payload]: _rm, ...rest } = state;
      return { ...rest };
    }
  }
}
