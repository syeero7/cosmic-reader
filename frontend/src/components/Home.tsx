import {
  AddComicBook,
  DeleteComic,
  GenerateULID,
  OpenCBZByID,
  OpenCBZByPath,
  SelectAnyComic,
  SelectOnlyCBZ,
} from "@wails/go/main/App";
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
    const path = await SelectAnyComic();
    if (!path) return;

    const id = await GenerateULID();
    dispatch({ type: "add_comics", payload: { [id]: "Loading..." } });
    AddComicBook(id, path).then((v) => dispatch({ type: "add_comics", payload: { [id]: v } }));
  };

  const deleteComic = (id: string) => {
    return async () => {
      await DeleteComic(id);
      dispatch({ type: "delete_comic", payload: id });
    };
  };

  const openComic = (id: string) => {
    return async () => {
      const cbz = await OpenCBZByID(id);
      router.navigateTo({ name: "comic-view", data: cbz });
    };
  };

  const openCBZ = async () => {
    const path = await SelectOnlyCBZ();
    if (!path) return;
    const cbz = await OpenCBZByPath(path);
    router.navigateTo({ name: "comic-view", data: cbz });
  };

  const archives = !filterQuery
    ? Object.entries(comics)
    : Object.entries(comics).filter(([_k, title]) =>
        title.toLowerCase().includes(filterQuery.toLowerCase()),
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
        {archives.map(([id, title]) => (
          <ComicCard
            key={id}
            comic={{ id, title }}
            deleteFn={deleteComic(id)}
            navigateFn={openComic(id)}
          />
        ))}
      </div>
    </main>
  );
}

type ComicCardProps = {
  comic: { id: string; title: string };
  deleteFn: () => void;
  navigateFn: () => void;
};

function ComicCard({ comic, deleteFn, navigateFn }: ComicCardProps) {
  const temporary = comic.title.startsWith("Loading");

  return (
    <article className={`comic-card${temporary ? " loading" : ""}`}>
      <button onClick={deleteFn} title={`delete ${comic.title}`} disabled={temporary}>
        <DeleteSVG />
      </button>

      <img
        src={`/thumbnails/${comic.id}`}
        alt={`${comic.title} cover`}
        key={comic.title}
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
