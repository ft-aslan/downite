import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Link, createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/")({
  component: HomePage,
});
export default function HomePage() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <Button onClick={() => setCount((count) => count + 1)}>{count}</Button>
      <Link to="/torrent">Torrent</Link>
    </div>
  );
}
