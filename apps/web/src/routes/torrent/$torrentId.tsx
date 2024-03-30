import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/torrent/$torrentId")({
  component: () => <div>Hello /torrent/$torrentId!</div>,
})
