import type { main } from "@wails/go/models";
import type { TargetedInputEvent } from "preact";
import { type Dispatch, type StateUpdater, useEffect, useRef, useState } from "preact/hooks";
import { useComics } from "./ComicProvider";
import { useRouter } from "./RouterProvider";
import { useShortcut } from "./ShortcutProvider";
import {
  HomeSVG,
  LockSVG,
  NextSVG,
  PreviousSVG,
  RotateLeftSVG,
  RotateRightSVG,
  UnlockSVG,
} from "./SVG";

type ImageOrientation = 0 | 90 | -90 | 180;

type ComicPageProps = {
  comicId: string;
  tempComic?: Pick<main.Archive, "pageCount">;
};

const getInitialPageState = (pageCount?: number) => {
  return { prev: 0, current: pageCount ? 1 : 0, next: pageCount ? Math.min(2, pageCount) : 0 };
};

export function ComicPage({ comicId, tempComic }: ComicPageProps) {
  const router = useRouter();
  const { comics } = useComics();
  const selectedComic = tempComic ? tempComic : comics[comicId];
  const [pages, setPages] = useState(() => getInitialPageState(selectedComic.pageCount));
  const [orientation, setOrientation] = useState<ImageOrientation>(0);

  if (!selectedComic) {
    router.navigateTo("home");
    return;
  }

  const landscape = [-90, 90].includes(orientation);

  const styles = {
    "--comic-page-orientation": `rotate(${orientation}deg)`,
    "--comic-page-width": `100${landscape ? "vh" : "vw"}`,
    "--comic-page-height": `100${landscape ? "vw" : "vh"}`,
    "--comic-page-url": `url(/comics/${comicId}/pages/${pages.current}?temp=${!!tempComic})`,
  };

  return (
    <main className="comic-page">
      <div className="page-container" style={styles}>
        <ComicMenu
          pages={pages}
          setPages={setPages}
          setOrientation={setOrientation}
          pageCount={selectedComic.pageCount}
          navigateToHome={() => router.navigateTo("home")}
        />
      </div>
    </main>
  );
}

type MenuProps = {
  pages: ReturnType<typeof getInitialPageState>;
  setPages: Dispatch<StateUpdater<ReturnType<typeof getInitialPageState>>>;
  setOrientation: Dispatch<StateUpdater<ImageOrientation>>;
  navigateToHome: () => void;
  pageCount: number;
};

function ComicMenu({ pageCount, pages, setPages, setOrientation, navigateToHome }: MenuProps) {
  const [lockState, setLockState] = useState<"lock" | "unlock">("unlock");
  const [pageInput, setPageInput] = useState(1);
  const timerRef = useRef<NodeJS.Timeout | undefined>();

  const toNextPage = () => {
    if (pages.current === pageCount) return;
    setPages((p) => ({ prev: p.current, current: p.next, next: p.next + 1 }));
    setPageInput(pages.next);
  };

  const toPreviousPage = () => {
    if (pages.current === 1) return;
    setPages((p) => ({ prev: p.prev - 1, current: p.prev, next: p.current }));
    setPageInput(pages.prev);
  };

  useShortcut("ArrowUp", toPreviousPage);
  useShortcut("ArrowRight", toNextPage);
  useShortcut("ArrowDown", toNextPage);
  useShortcut("ArrowLeft", toPreviousPage);

  useShortcut("Alt+ArrowUp", () => setOrientation(180)); // upside down
  useShortcut("Alt+ArrowRight", () => setOrientation(-90)); // counter clockwise
  useShortcut("Alt+ArrowDown", () => setOrientation(0)); // normal
  useShortcut("Alt+ArrowLeft", () => setOrientation(90)); // clockwise

  useMouseWheel(toPreviousPage, toNextPage);

  const toAnyPage = (e: TargetedInputEvent<HTMLInputElement>) => {
    const pageN = Number((e.target as HTMLInputElement).value);
    if (Number.isNaN(pageN) || pageN <= 0 || pageN > pageCount) return;
    setPageInput(pageN);
  };

  useEffect(() => {
    timerRef.current = setTimeout(() => {
      setPages({
        prev: pageInput - 1,
        current: pageInput,
        next: pageInput === pageCount ? 0 : pageInput + 1,
      });
    }, 200);

    return () => clearTimeout(timerRef.current);
  }, [setPages, pageInput]);

  const toggleMenuVisibility = () => setLockState((l) => (l === "lock" ? "unlock" : "lock"));
  useShortcut("Alt+l", toggleMenuVisibility);

  const rotateLeft = () => setOrientation((o) => getOrientation(o, "L"));
  const rotateRight = () => setOrientation((o) => getOrientation(o, "R"));

  return (
    <menu>
      <div className="page-menu" style={{ "--page-menu-visibility": lockState === "lock" ? 1 : 0 }}>
        <div className="page-input">
          <input type="range" min={1} value={pageInput} max={pageCount} onChange={toAnyPage} />
          <span>{pageInput}</span>
        </div>

        <div className="menu-btns">
          <button onClick={navigateToHome} class="to-home-btn" title="go back">
            <HomeSVG />
          </button>
          <button onClick={rotateLeft} title="rotate left">
            <RotateLeftSVG />
          </button>
          <button onClick={rotateRight} title="rotate right">
            <RotateRightSVG />
          </button>
          <button title={lockState} onClick={toggleMenuVisibility}>
            {lockState === "lock" ? <LockSVG /> : <UnlockSVG />}
          </button>
          <button onClick={toPreviousPage} title="previous page">
            <PreviousSVG />
          </button>
          <button onClick={toNextPage} title="next page">
            <NextSVG />
          </button>
        </div>
      </div>
    </menu>
  );
}

function useMouseWheel(onWheelUp: () => void, onWheelDown: () => void) {
  const throttleRef = useRef(0);

  useEffect(() => {
    const handleWheel = (e: WheelEvent) => {
      const now = Date.now();
      const delay = 300;

      if (now - throttleRef.current < delay) return;
      if (e.deltaY < 0) {
        onWheelUp();
      } else {
        onWheelDown();
      }

      throttleRef.current = now;
    };

    window.addEventListener("wheel", handleWheel);
    return () => window.removeEventListener("wheel", handleWheel);
  }, [onWheelUp, onWheelDown]);
}

function getOrientation(orientation: ImageOrientation, direction: "L" | "R"): ImageOrientation {
  switch (orientation) {
    case 0:
      return direction === "L" ? -90 : 90;
    case 90:
      return direction === "L" ? 0 : 180;
    case -90:
      return direction === "L" ? 180 : 0;
    case 180:
      return direction === "L" ? 90 : -90;
  }
}
