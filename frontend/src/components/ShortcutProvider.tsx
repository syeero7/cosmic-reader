import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useRef,
} from "preact/compat";

type ArrowKeys = "ArrowDown" | "ArrowUp" | "ArrowLeft" | "ArrowRight";
type KeyCombo =
  | "Control+Shift+f"
  | "Control+q"
  | "Alt+l"
  | "Alt+k"
  | ArrowKeys
  | `Alt+${ArrowKeys}`;
type Keys = "Control" | "Shift" | "Alt" | (string & {});
type LowercasedCombo = Lowercase<KeyCombo>;

type ShortcutCtx = {
  register: (combo: LowercasedCombo, fn: () => void) => void;
  unregister: (combo: LowercasedCombo) => void;
};

const ShortcutContext = createContext<ShortcutCtx | null>(null);

export function useShortcut(combo: KeyCombo, fn: () => void) {
  const ctx = useContext(ShortcutContext);
  if (!ctx) throw new Error("useShortcut hook must be used within a child of ShortcutProvider");

  useEffect(() => {
    const lowercased = combo.toLowerCase() as LowercasedCombo;
    ctx.register(lowercased, fn);
    return () => ctx.unregister(lowercased);
  }, [fn]);
}

export function ShortcutProvider({ children }: PropsWithChildren) {
  const keyMapRef = useRef<Map<LowercasedCombo, () => void>>(new Map());
  const throttleRef = useRef(0);

  useEffect(() => {
    const handleKeydown = (e: KeyboardEvent) => {
      const callbackFn = keyMapRef.current.get(buildKeyCombo(e));
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

function buildKeyCombo(e: KeyboardEvent) {
  const keys: Keys[] = [];
  if (e.ctrlKey) keys.push("Control");
  if (e.shiftKey) keys.push("Shift");
  if (e.altKey) keys.push("Alt");
  keys.push(e.key);

  return keys.join("+").toLowerCase() as LowercasedCombo;
}
