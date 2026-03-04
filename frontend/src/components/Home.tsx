import { AddComicBook, DeleteComic, OpenCBZFile, SelectFile } from "@wails/go/main/App";
import type { main } from "@wails/go/models";
import type { TargetedInputEvent } from "preact";
import { useRef, useState } from "preact/compat";
import fallbackImg from "@/assets/cosmic_fallback.webp";
import { useComics } from "./ComicProvider";
import { useRouter } from "./RouterProvider";
import { AddSVG, DeleteSVG, KBDShortcutSVG, OpenFileSVG } from "./SVG";

export function Home() {
  const [filterQuery, setFilterQuery] = useState("");
  const { comics, dispatch } = useComics();
  const router = useRouter();
  const timerRef = useRef<NodeJS.Timeout | undefined>();

  const filterComics = (e: TargetedInputEvent<HTMLInputElement>) => {
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      const target = e.target as HTMLInputElement;
      setFilterQuery(target.value.trim());
    }, 300);
  };

  const addComic = async () => {
    const path = await SelectFile();
    if (!path) return;

    const id = crypto.randomUUID();
    const tmp: main.Archive = { title: "Loading...", pageCount: 0, thumbnail: "" };
    dispatch({ type: "add_comics", payload: { [id]: tmp } });
    AddComicBook(id, path).then((v) => dispatch({ type: "add_comics", payload: { [id]: v } }));
  };

  const deleteComic = (id: string) => {
    return async () => {
      await DeleteComic(id);
      dispatch({ type: "delete_comic", payload: id as ReturnType<typeof crypto.randomUUID> });
    };
  };

  const openComic = (id: string) => {
    return () => {
      router.navigateTo(`comic-id: ${id as ReturnType<typeof crypto.randomUUID>}`);
    };
  };

  const openCBZ = async () => {
    const pageCount = await OpenCBZFile();
    if (pageCount === 0) return;
    router.navigateTo(`temp-comic: ${crypto.randomUUID()},${pageCount}`);
  };

  const archives = !filterQuery
    ? Object.entries(comics)
    : Object.entries(comics).filter(([_k, v]) =>
        v.title.toLowerCase().includes(filterQuery.toLowerCase()),
      );

  return (
    <main className="homepage">
      <header>
        <h1>Library</h1>
        <input type="search" onChange={filterComics} value={filterQuery} placeholder="Search..." />
        <button onClick={addComic} title="add comic book">
          <AddSVG />
        </button>

        <button onClick={openCBZ} title="open comic book (.cbz)">
          <OpenFileSVG />
        </button>

        <button onClick={() => router.openModal("kbd-shortcuts")} title="view keybord shortcuts">
          <KBDShortcutSVG />
        </button>
      </header>

      <div className="comic-grid">
        {archives.map(([k, v]) => (
          <ComicCard
            key={k}
            comic={{ ...v, id: k }}
            deleteFn={deleteComic(k)}
            navigateFn={openComic(k)}
          />
        ))}
      </div>
    </main>
  );
}

type ComicCardProps = {
  comic: main.Archive & { id: string };
  deleteFn: () => void;
  navigateFn: () => void;
};

function ComicCard({ comic, deleteFn, navigateFn }: ComicCardProps) {
  const temporary = comic.title.startsWith("Loading") && comic.pageCount === 0;

  return (
    <article className={`comic-card${temporary ? " loading" : ""}`}>
      <button onClick={deleteFn} title={`delete ${comic.title}`} disabled={temporary}>
        <DeleteSVG />
      </button>

      <img
        src={`/thumbnails/${comic.thumbnail}`}
        alt={`${comic.title} cover`}
        onError={(e) => {
          (e.target as HTMLImageElement).src = fallbackImg;
        }}
      />
      <a
        title={comic.title}
        href={`/comics/${comic.id}`}
        onClick={(e) => {
          e.preventDefault();
          if (temporary) return;
          navigateFn();
        }}
      >
        {comic.title}
      </a>
    </article>
  );
}
