import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/torrents/")({
  component: () => TorrentsPage(),
  // loader: ({ context: { queryClient } }) => {
  //   // queryClient.ensureQueryData(getTorrentsQueryOptions())
  // },
})

import { Button } from "@/components/ui/button"
import { AddTorrentDialog } from "./-components/AddTorrentDialog"
import { PlusCircle } from "lucide-react"
import { TorrentsTable } from "./-components/TorrentsTable"
import { useSocketClient } from "@/api"
import React, { useCallback, useMemo } from "react"
import { components } from "@/api/v1"

function TorrentsPage() {
  const { sendJsonMessage, lastJsonMessage } = useSocketClient<{torrents: components["schemas"]["Torrent"][]}>("/torrent-ws");

  const data = useMemo(() => {
    return { ...lastJsonMessage };
  }, [lastJsonMessage])

  const refetch = useCallback(() => {
    sendJsonMessage({
      method: "GET"
    })
  }, [sendJsonMessage])

  React.useEffect(() => {
    const tableUpdateInterval = setInterval(() => refetch(), 1000)
    return () => clearInterval(tableUpdateInterval)
  }, [])

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
      <TorrentsTable torrents={data?.torrents || []} />
    </div>
  )
}
