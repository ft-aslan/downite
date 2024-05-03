import { Button } from "@/components/ui/button"
import { Link, createFileRoute } from "@tanstack/react-router"
import { AddTorrentDialog } from "./-components/AddTorrentDialog"
import { PlusCircle } from "lucide-react"

export const Route = createFileRoute("/")({
  component: HomePage,
})
export default function HomePage() {
  return (
    <div>
      <AddTorrentDialog type="magnet">
        <Button variant="default" className="gap-1">
          <PlusCircle className="h-3.5 w-3.5" />
          <span className="sm:whitespace-nowrap">Add Magnet</span>
        </Button>
      </AddTorrentDialog>
      <AddTorrentDialog type="file">
        <Button variant="default" className="gap-1">
          <PlusCircle className="h-3.5 w-3.5" />
          <span className="sm:whitespace-nowrap">Add Torrent</span>
        </Button>
      </AddTorrentDialog>
      <Link to="/torrent">Torrent</Link>
    </div>
  )
}
