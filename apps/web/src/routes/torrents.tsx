import { Button } from "@/components/ui/button"

import { Outlet, createFileRoute } from "@tanstack/react-router"
import { Share } from "lucide-react"

export const Route = createFileRoute("/torrents")({
  component: () => TorrentsRoot(),
})
function TorrentsRoot() {
  return (
    <div>
      <header className="bg-background sticky top-0 z-10 flex h-[57px] items-center gap-1 border-b px-4">
        <h1 className="text-xl font-semibold">Torrents</h1>
      </header>
      <main className="p-4">
        <Outlet />
      </main>
    </div>
  )
}
