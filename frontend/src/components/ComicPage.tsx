import type { main } from "@wails/go/models";
import type { TargetedInputEvent } from "preact";
import { type Dispatch, type StateUpdater, useRef, useState } from "preact/hooks";
import { useComics } from "./ComicProvider";
import { useRouter } from "./RouterProvider";

const getInitialPageState = (comic?: main.ArchiveInfo) => {
  return { prev: 0, current: comic ? 1 : 0, next: comic ? Math.min(2, comic.pageCount) : 0 };
};

export function ComicPage({ comicId }: { comicId: string }) {
  const router = useRouter();
  const { comics } = useComics();
  const selectedComic = comics.find((c) => c.id === comicId);
  const [pages, setPages] = useState(() => getInitialPageState(selectedComic));

  if (!selectedComic) {
    router.navigateTo("home");
    return;
  }

  return (
    <main>
      <img
        src={`/comics/${selectedComic.id}/pages/${pages.current}`}
        alt={`page ${pages.current}`}
      />
      <ComicMenu
        pages={pages}
        setPages={setPages}
        pageCount={selectedComic.pageCount}
        navigateToHome={() => router.navigateTo("home")}
      />
    </main>
  );
}

type MenuProps = {
  pages: ReturnType<typeof getInitialPageState>;
  setPages: Dispatch<StateUpdater<ReturnType<typeof getInitialPageState>>>;
  navigateToHome: () => void;
  pageCount: number;
};

function ComicMenu({ pageCount, pages, setPages, navigateToHome }: MenuProps) {
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
      setPages({ prev: pageN - 1, current: pageN, next: pageN === pageCount ? 0 : pageN + 1 });
    }, 300);
  };

  // TODO: add screen rotate (https://developer.mozilla.org/en-US/docs/Web/API/ScreenOrientation/lock) n menu lock

  return (
    <menu>
      <button onClick={navigateToHome}>Home</button>
      <input type="range" min={1} defaultValue={1} max={pageCount} onChange={toAnyPage} />
      <div>
        <button>rotate left</button>
        <button>rotate right</button>
      </div>

      <button>lock menu</button>

      <div>
        <button onClick={toPreviousPage}>prev page</button>
        <button onClick={toNextPage}>next page</button>
      </div>
    </menu>
  );
}
