import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Link, createFileRoute } from "@tanstack/react-router"
import { AddTorrentRename } from "./-components/AddTorrentDialog"

export const Route = createFileRoute("/")({
  component: HomePage,
})
export default function HomePage() {
  return (
    <div>
      <AddTorrentRename></AddTorrentRename>
      <Link to="/torrent">Torrent</Link>
    </div>
  )
}
