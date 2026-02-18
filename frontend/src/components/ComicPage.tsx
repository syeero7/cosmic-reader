import type { main } from "@wails/go/models";
import type { TargetedInputEvent } from "preact";
import { type Dispatch, type StateUpdater, useRef, useState } from "preact/hooks";
import { useComics } from "./ComicProvider";
import { useRouter } from "./RouterProvider";

type ImageOrientation = 0 | 90 | -90 | 180;

const getInitialPageState = (comic?: main.ArchiveInfo) => {
  return { prev: 0, current: comic ? 1 : 0, next: comic ? Math.min(2, comic.pageCount) : 0 };
};

export function ComicPage({ comicId }: { comicId: string }) {
  const router = useRouter();
  const { comics } = useComics();
  const selectedComic = comics.find((c) => c.id === comicId);
  const [pages, setPages] = useState(() => getInitialPageState(selectedComic));
  const [orientation, setOrientation] = useState<ImageOrientation>(0);

  if (!selectedComic) {
    router.navigateTo("home");
    return;
  }

  const landscape = [-90, 90].some((v) => v === orientation);

  const styles = {
    "--comic-page-orientation": `rotate(${orientation}deg)`,
    "--comic-page-width": `100${landscape ? "vh" : "vw"}`,
    "--comic-page-height": `100${landscape ? "vw" : "vh"}`,
    "--comic-page-url": `url(/comics/${selectedComic.id}/pages/${pages.current})`,
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
  const timerRef = useRef<NodeJS.Timeout | undefined>();

  const toNextPage = () => {
    if (pages.current === pageCount) return;
    setPages((p) => ({ prev: p.current, current: p.next, next: p.next + 1 }));
  };

  const toPreviousPage = () => {
    if (pages.current === 1) return;
    setPages((p) => ({ prev: p.prev - 1, current: p.prev, next: p.current }));
  };

  const toAnyPage = (e: TargetedInputEvent<HTMLInputElement>) => {
    clearTimeout(timerRef.current);
    const pageN = Number((e.target as HTMLInputElement).value);
    if (Number.isNaN(pageN) || pageN <= 0 || pageN > pageCount) return;

    timerRef.current = setTimeout(() => {
      setPages({
        prev: pageN - 1,
        current: pageN,
        next: pageN === pageCount ? 0 : pageN + 1,
      });
    }, 300);
  };

  const rotateLeft = () => setOrientation((o) => getOrientation(o, "L"));
  const rotateRight = () => setOrientation((o) => getOrientation(o, "R"));

  return (
    <menu>
      <div className="page-menu">
        <div className="page-input">
          <input type="range" min={1} defaultValue={1} max={pageCount} onChange={toAnyPage} />
          <span>{pages.current}</span>
        </div>

        <div className="menu-btns">
          <button onClick={navigateToHome} class="to-home-btn" title="go back">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
              <path d="M240-200h120v-240h240v240h120v-360L480-740 240-560v360Zm-80 80v-480l320-240 320 240v480H520v-240h-80v240H160Zm320-350Z" />
            </svg>
          </button>
          <button onClick={rotateLeft} title="rotate left">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
              <path d="M440-80q-50-5-96-24.5T256-156l56-58q29 21 61.5 34t66.5 18v82Zm80 0v-82q104-15 172-93.5T760-438q0-117-81.5-198.5T480-718h-8l64 64-56 56-160-160 160-160 56 58-62 62h6q75 0 140.5 28.5t114 77q48.5 48.5 77 114T840-438q0 137-91 238.5T520-80ZM198-214q-32-42-51.5-88T122-398h82q5 34 18 66.5t34 61.5l-58 56Zm-76-264q6-51 25-98t51-86l58 56q-21 29-34 61.5T204-478h-82Z" />
            </svg>
          </button>
          <button onClick={rotateRight} title="rotate right">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
              <path d="M522-80v-82q34-5 66.5-18t61.5-34l56 58q-42 32-88 51.5T522-80Zm-80 0Q304-98 213-199.5T122-438q0-75 28.5-140.5t77-114q48.5-48.5 114-77T482-798h6l-62-62 56-58 160 160-160 160-56-56 64-64h-8q-117 0-198.5 81.5T202-438q0 104 68 182.5T442-162v82Zm322-134-58-56q21-29 34-61.5t18-66.5h82q-5 50-24.5 96T764-214Zm76-264h-82q-5-34-18-66.5T706-606l58-56q32 39 51 86t25 98Z" />
            </svg>
          </button>
          <button title="lock">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
              <path d="M240-80q-33 0-56.5-23.5T160-160v-400q0-33 23.5-56.5T240-640h40v-80q0-83 58.5-141.5T480-920q83 0 141.5 58.5T680-720v80h40q33 0 56.5 23.5T800-560v400q0 33-23.5 56.5T720-80H240Zm0-80h480v-400H240v400Zm296.5-143.5Q560-327 560-360t-23.5-56.5Q513-440 480-440t-56.5 23.5Q400-393 400-360t23.5 56.5Q447-280 480-280t56.5-23.5ZM360-640h240v-80q0-50-35-85t-85-35q-50 0-85 35t-35 85v80ZM240-160v-400 400Z" />
            </svg>
          </button>
          <button onClick={toPreviousPage} title="previous page">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
              <path d="M560-240 320-480l240-240 56 56-184 184 184 184-56 56Z" />
            </svg>
          </button>
          <button onClick={toNextPage} title="next page">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
              <path d="M504-480 320-664l56-56 240 240-240 240-56-56 184-184Z" />
            </svg>
          </button>
        </div>
      </div>
    </menu>
  );
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
