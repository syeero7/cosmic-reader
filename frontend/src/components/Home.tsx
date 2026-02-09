import { AddComicBook, DeleteComic, SelectFile } from "@wails/go/main/App";
import type { main } from "@wails/go/models";
import type { TargetedInputEvent } from "preact";
import { useState } from "preact/compat";
import { useComics } from "./ComicProvider";

export function Home() {
  const [filterQuery, setFilterQuery] = useState("");
  const { comics, dispatch } = useComics();

  const filterComics = () => {
    let timer: NodeJS.Timeout | undefined;
    return (e: TargetedInputEvent<HTMLInputElement>) => {
      clearTimeout(timer);
      timer = setTimeout(() => {
        const target = e.target as HTMLInputElement;
        setFilterQuery(target.value.trim());
      }, 300);
    };
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
      dispatch({
        type: "delete_comic",
        payload: id as ReturnType<typeof crypto.randomUUID>,
      });
    };
  };

  const archives = !filterQuery
    ? comics
    : comics.filter((c) => c.title.toLowerCase().includes(filterQuery.toLowerCase()));

  return (
    <main>
      <div>
        <button onClick={addComic}>add</button>
        <div>
          <span>search</span>
          <input type="text" onChange={filterComics()} value={filterQuery} />
        </div>
      </div>

      <div>
        {archives.map((c) => (
          <ComicCard key={c.id} comic={c} deleteFn={deleteComic(c.id)} />
        ))}
      </div>
    </main>
  );
}

function ComicCard({ comic, deleteFn }: { comic: main.ArchiveInfo; deleteFn: () => void }) {
  return (
    <article>
      <button onClick={deleteFn}>delete</button>
      <img src={`/thumbnails/${comic.thumbnail}`} alt={`${comic.title} cover`} />
      <p>{comic.title}</p>
      {/* 10ch max width title */}
    </article>
  );
}
