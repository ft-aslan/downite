import { createFileRoute } from "@tanstack/react-router"
export const Route = createFileRoute("/torrents/")({
  component: () => TorrentsPage(),
  loader: ({ context: { queryClient } }) => {
    queryClient.ensureQueryData(getTorrentsQueryOptions())
  },
})

import { Button } from "@/components/ui/button"
import { AddTorrentDialog } from "./-components/AddTorrentDialog"
import { PlusCircle } from "lucide-react"
import { TorrentsTable } from "./-components/TorrentsTable"
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query"
import { client } from "@/api"
import React from "react"
const getTorrentsQueryOptions = () =>
  queryOptions({
    queryKey: ["torrents"],
    queryFn: () => client.GET("/torrent"),
  })
function TorrentsPage() {
  const {
    data: { data },
    refetch,
  } = useSuspenseQuery(getTorrentsQueryOptions())
  React.useEffect(() => {
    const tableUpdateInterval = setInterval(() => refetch(), 1000)
    return () => clearInterval(tableUpdateInterval)
  }, [])

  if (!data) return null

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
      <TorrentsTable torrents={data?.torrents} />
    </div>
  )
}
