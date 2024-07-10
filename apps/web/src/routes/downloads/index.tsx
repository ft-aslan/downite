import { createFileRoute } from "@tanstack/react-router"
export const Route = createFileRoute("/downloads/")({
  component: () => DownloadsPage(),
  loader: ({ context: { queryClient } }) => {
    queryClient.ensureQueryData(getDownloadsQueryOptions())
  },
})

import { Button } from "@/components/ui/button"
import { PlusCircle } from "lucide-react"
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query"
import { client } from "@/api"
import React from "react"
import { AddDownloadDialog } from "./-components/AddDownloadDialog"
import { DownloadsTable } from "./-components/DownloadsTable"

const getDownloadsQueryOptions = () =>
  queryOptions({
    queryKey: ["downloads"],
    queryFn: async () => {
      try {
        const { data } = await client.GET("/download")
        return data
      } catch (error) {
        return { downloads: [] }
      }
    },
  })
function DownloadsPage() {
  const { data, refetch } = useSuspenseQuery(getDownloadsQueryOptions())
  React.useEffect(() => {
    const tableUpdateInterval = setInterval(() => refetch(), 1000)
    return () => clearInterval(tableUpdateInterval)
  }, [])

  return (
    <div>
      <div className="flex items-center gap-2">
        <AddDownloadDialog>
          <Button variant="default" className="gap-1">
            <PlusCircle className="h-3.5 w-3.5" />
            <span className="sm:whitespace-nowrap">Add Download</span>
          </Button>
        </AddDownloadDialog>
      </div>
      <DownloadsTable downloads={data || []} />
    </div>
  )
}
