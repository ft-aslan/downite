import { client } from "@/api"
import { Outlet, createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/torrent/$infohash")({
  component: () => TorrentRoot(),
  loader: ({ context: { queryClient }, params: { infohash } }) => {
    queryClient.ensureQueryData({
      queryKey: ["torrent", infohash],
      queryFn: () =>
        client.GET("/torrent/{infohash}", {
          params: {
            infohash,
          },
        }),
    })
  },
})

function TorrentRoot() {
  return (
    <div>
      <header className="bg-background sticky top-0 z-10 flex h-[57px] items-center gap-1 border-b px-4">
        <h1 className="text-xl font-semibold">Torrent</h1>
      </header>
      <main className="p-4"></main>
    </div>
  )
}
