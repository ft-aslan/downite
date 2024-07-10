import { Outlet, createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/downloads")({
  component: () => DownloadsRoot(),
})

function DownloadsRoot() {
  return (
    <div>
      <header className="bg-background sticky top-0 z-10 flex h-[57px] items-center gap-1 border-b px-4">
        <h1 className="text-xl font-semibold">Downloads</h1>
      </header>
      <main className="p-4">
        <Outlet />
      </main>
    </div>
  )
}
