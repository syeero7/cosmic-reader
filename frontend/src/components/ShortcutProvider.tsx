import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useRef,
} from "preact/compat";

type ArrowKeys = "ArrowDown" | "ArrowUp" | "ArrowLeft" | "ArrowRight";
type KeyCombo = "Control+Shift+f" | ArrowKeys | `Alt+${ArrowKeys}`;
type Keys = "Control" | "Shift" | "Alt" | (string & {});

type ShortcutCtx = {
  register: (combo: KeyCombo, fn: () => void) => void;
  unregister: (combo: KeyCombo) => void;
};

const ShortcutContext = createContext<ShortcutCtx | null>(null);

export function useShortcut(combo: KeyCombo, fn: () => void) {
  const ctx = useContext(ShortcutContext);
  if (!ctx) throw new Error("useShortcut hook must be used within a child of ShortcutProvider");

  useEffect(() => {
    ctx.register(combo, fn);
    return () => ctx.unregister(combo);
  }, [fn]);
}

export function ShortcutProvider({ children }: PropsWithChildren) {
  const keyMapRef = useRef<Map<KeyCombo, () => void>>(new Map());
  const throttleRef = useRef(0);

  useEffect(() => {
    const handleKeydown = (e: KeyboardEvent) => {
      const keys: Keys[] = [];
      if (e.ctrlKey) keys.push("Control");
      if (e.shiftKey) keys.push("Shift");
      if (e.altKey) keys.push("Alt");
      keys.push(e.key);

      const callbackFn = keyMapRef.current.get(keys.join("+") as KeyCombo);
      if (!callbackFn) return;
      e.preventDefault();

      if (!e.repeat) {
        throttleRef.current = Date.now();
        callbackFn();
        return;
      }

      const delay = 220;
      const now = Date.now();
      if (now - throttleRef.current >= delay) {
        throttleRef.current = now;
        callbackFn();
      }
    };

    window.addEventListener("keydown", handleKeydown);
    return () => window.removeEventListener("keydown", handleKeydown);
  }, []);

  const register: ShortcutCtx["register"] = (combo, fn) => keyMapRef.current.set(combo, fn);
  const unregister: ShortcutCtx["unregister"] = (combo) => keyMapRef.current.delete(combo);

  return (
    <ShortcutContext.Provider value={{ register, unregister }}>{children}</ShortcutContext.Provider>
  );
}
