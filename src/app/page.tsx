import { useState } from "react";
import { Button } from "@/components/ui/button";

export default function HomePage() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <Button onClick={() => setCount((count) => count + 1)}>{count}</Button>
    </div>
  );
}
