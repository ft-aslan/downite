import { Button } from "@/components/ui/button"
import { Link, createFileRoute } from "@tanstack/react-router"
import { AddTorrentDialog } from "./-components/AddTorrentDialog"
import { PlusCircle } from "lucide-react"
import { TorrentsTable } from "./-components/TorrentsTable"
export const Route = createFileRoute("/torrents/")({
  component: () => TorrentsPage(),
})
function TorrentsPage() {
  return (
    <div>
      <div className="flex items-center gap-2">
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
      </div>
      <TorrentsTable />
    </div>
  )
}
