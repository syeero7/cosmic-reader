import { useEffect, useRef } from "preact/hooks";
import { useRouter } from "./RouterProvider";

export function KbdShortcutModal() {
  const [modalRef, closeModal] = useModal("kbd-shortcuts");

  return (
    <dialog ref={modalRef}>
      <div className="kbd-modal">
        <button onClick={closeModal} title="Close">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
            <path d="m336-280-56-56 144-144-144-143 56-56 144 144 143-144 56 56-144 143 144 144-56 56-143-144-144 144Z" />
          </svg>
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
              <tr>
                <td>
                  <kbd>Ctrl</kbd> + <kbd>q</kbd>
                </td>
                <td>Quite Cosmic Reader</td>
              </tr>

              <tr>
                <td>
                  <kbd>Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>f</kbd>
                </td>
                <td>Toggle Fullscreen</td>
              </tr>

              <tr>
                <td>
                  <kbd>Alt</kbd> + <kbd>l</kbd>
                </td>
                <td>Toggle Page Menu Visibility</td>
              </tr>

              <tr>
                <td>
                  <kbd>Alt</kbd> + <kbd>k</kbd>
                </td>
                <td>Toggle Keyboard Shortcuts Modal</td>
              </tr>

              <tr>
                <td>
                  <kbd>Arrow Up</kbd>, <kbd>Arrow Left</kbd>
                </td>
                <td>Go to Previous Page</td>
              </tr>

              <tr>
                <td>
                  <kbd>Arrow Down</kbd>, <kbd>Arrow Right</kbd>
                </td>
                <td>Go to Next Page</td>
              </tr>

              <tr>
                <td>
                  <kbd>Alt</kbd> + <kbd>Arrow Down</kbd>
                </td>
                <td>Rotate Normal 0&deg;</td>
              </tr>

              <tr>
                <td>
                  <kbd>Alt</kbd> + <kbd>Arrow Up</kbd>
                </td>
                <td>Rotate Upside Down 180&deg;</td>
              </tr>

              <tr>
                <td>
                  <kbd>Alt</kbd> + <kbd>Arrow Right</kbd>
                </td>
                <td>Rotate Clockwise 90&deg;</td>
              </tr>

              <tr>
                <td>
                  <kbd>Alt</kbd> + <kbd>Arrow Left</kbd>
                </td>
                <td>Rotate Counter Clockwise -90&deg;</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </dialog>
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
