import { AddComicBook, DeleteComic, SelectFile } from "@wails/go/main/App";
import type { main } from "@wails/go/models";
import type { TargetedInputEvent } from "preact";
import { useRef, useState } from "preact/compat";
import addImg from "@/assets/add-svgrepo-com.svg";
import deleteImg from "@/assets/delete-filled-svgrepo-com.svg";
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
          <img alt="add" src={addImg} width={36} height={36} />
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
        <img alt="delete" src={deleteImg} width={24} height={24} />
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
