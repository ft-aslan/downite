import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/torrent/")({
  component: () => <div>Hello /torrent/!</div>,
})
