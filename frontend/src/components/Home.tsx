import { AddComicBook, DeleteComic, SelectFile } from "@wails/go/main/App";
import type { main } from "@wails/go/models";
import type { TargetedInputEvent } from "preact";
import { useRef, useState } from "preact/compat";
import { useComics } from "./ComicProvider";
import { useRouter } from "./RouterProvider";

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
    const cbook = await AddComicBook(id, path);
    dispatch({ type: "add_comics", payload: [{ id, ...cbook }] });
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

  const archives = !filterQuery
    ? comics
    : comics.filter((c) => c.title.toLowerCase().includes(filterQuery.toLowerCase()));

  return (
    <main className="homepage">
      <header>
        <h1>Library</h1>
        <input type="search" onChange={filterComics} value={filterQuery} placeholder="Search..." />
        <button onClick={addComic} title="add comic book">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
            <path d="M440-440H200v-80h240v-240h80v240h240v80H520v240h-80v-240Z" />
          </svg>
        </button>

        <button onClick={() => router.openModal("kbd-shortcuts")}>
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
            <path d="M160-200q-33 0-56.5-23.5T80-280v-400q0-33 23.5-56.5T160-760h640q33 0 56.5 23.5T880-680v400q0 33-23.5 56.5T800-200H160Zm0-80h640v-400H160v400Zm160-40h320v-80H320v80ZM200-440h80v-80h-80v80Zm120 0h80v-80h-80v80Zm120 0h80v-80h-80v80Zm120 0h80v-80h-80v80Zm120 0h80v-80h-80v80ZM200-560h80v-80h-80v80Zm120 0h80v-80h-80v80Zm120 0h80v-80h-80v80Zm120 0h80v-80h-80v80Zm120 0h80v-80h-80v80ZM160-280v-400 400Z" />
          </svg>
        </button>
      </header>

      <div className="comic-grid">
        {archives.map((c) => (
          <ComicCard
            key={c.id}
            comic={c}
            deleteFn={deleteComic(c.id)}
            navigateFn={openComic(c.id)}
          />
        ))}
      </div>
    </main>
  );
}

type ComicCardProps = {
  comic: main.ArchiveInfo;
  deleteFn: () => void;
  navigateFn: () => void;
};

function ComicCard({ comic, deleteFn, navigateFn }: ComicCardProps) {
  return (
    <article className="comic-card">
      <button onClick={deleteFn} title={`delete ${comic.title}`}>
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 -960 960 960">
          <path d="M280-120q-33 0-56.5-23.5T200-200v-520h-40v-80h200v-40h240v40h200v80h-40v520q0 33-23.5 56.5T680-120H280Zm80-160h80v-360h-80v360Zm160 0h80v-360h-80v360Z" />
        </svg>
      </button>

      <img src={`/thumbnails/${comic.thumbnail}`} alt={`${comic.title} cover`} />
      <a
        title={comic.title}
        href={`/comics/${comic.id}`}
        onClick={(e) => {
          e.preventDefault();
          navigateFn();
        }}
      >
        {comic.title}
      </a>
    </article>
  );
}
