import { useEffect, useRef } from "preact/hooks";
import { useRouter } from "./RouterProvider";
import type { ShortcutKeys } from "./ShortcutProvider";
import { CloseSVG } from "./SVG";

type ShortcutRow = {
  keys: (Exclude<ShortcutKeys, "Control"> | "Ctrl")[];
  description: string;
};

export function KbdShortcutModal() {
  const [modalRef, closeModal] = useModal("kbd-shortcuts");

  return (
    <dialog ref={modalRef}>
      <div className="kbd-modal">
        <button onClick={closeModal} title="Close">
          <CloseSVG />
        </button>

        <div>
          <table>
            <thead>
              <tr>
                <th>Shortcut</th>
                <th>Description</th>
              </tr>
            </thead>

            <tbody>
              {shortcuts.map((sr, i) => (
                <ShortcutTR key={i} row={sr} />
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </dialog>
  );
}

function ShortcutTR({ row }: { row: ShortcutRow }) {
  return (
    <tr>
      <td>
        {row.keys.map((k) => {
          if (k === "+") return " + ";
          if (k === ",") return ", ";

          const arrowPrefix = "Arrow";
          if (k.startsWith(arrowPrefix)) {
            k = `${arrowPrefix} ${k.slice(arrowPrefix.length)}`;
          }

          return <kbd>{k}</kbd>;
        })}
      </td>
      <td>{row.description}</td>
    </tr>
  );
}

function useModal(modal: ReturnType<typeof useRouter>["activeModal"]) {
  const { activeModal, closeModal } = useRouter();
  const modalRef = useRef<HTMLDialogElement | null>(null);

  useEffect(() => {
    if (!modalRef.current) return;
    if (activeModal === modal) {
      modalRef.current.showModal();
      return;
    }

    modalRef.current.close();
  }, [activeModal]);

  return [modalRef, closeModal] as const;
}

const shortcuts: ShortcutRow[] = [
  { keys: ["Ctrl", "+", "q"], description: "Quite Cosmic Reader" },
  { keys: ["Ctrl", "+", "Shift", "+", "f"], description: "Toggle Fullscreen" },
  { keys: ["Alt", "+", "l"], description: "Toggle Page Menu Visibility" },
  { keys: ["Alt", "+", "k"], description: "Toggle Keyboard Shortcuts Modal" },
  { keys: ["ArrowDown", ",", "ArrowRight"], description: "Go to Next Page" },
  { keys: ["ArrowUp", ",", "ArrowLeft"], description: "Go to Previous Page" },
  { keys: ["Alt", "+", "ArrowDown"], description: "Rotate Normal (0\u00B0)" },
  { keys: ["Alt", "+", "ArrowUp"], description: "Rotate Upside Down (180\u00B0)" },
  { keys: ["Alt", "+", "ArrowLeft"], description: "Rotate Clockwise (90\u00B0)" },
  { keys: ["Alt", "+", "ArrowRight"], description: "Rotate Counter Clockwise (-90\u00B0)" },
];
