import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/download/$id")({
  component: () => DownloadRoot(),
  loader: ({ context: { queryClient }, params: { id } }) => {
    queryClient.ensureQueryData(getDownloadQueryOptions(id))
  },
})

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { capitalizeFirstLetter } from "@/lib/utils"
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query"
import React from "react"
import { client } from "@/api"
import { Progress } from "@/components/ui/progress"

const getDownloadQueryOptions = (id: string) =>
  queryOptions({
    queryKey: ["download", id],
    queryFn: () =>
      client.GET("/download/{id}", {
        params: {
          path: {
            id,
          },
        },
      }),
  })
function DownloadRoot() {
  const { id } = Route.useParams()
  const {
    data: { data: download },
    refetch,
  } = useSuspenseQuery(getDownloadQueryOptions(id))
  React.useEffect(() => {
    const downloadUpdateInterval = setInterval(() => refetch(), 1000)
    return () => clearInterval(downloadUpdateInterval)
  }, [])

  if (!download) return null
  return (
    <Tabs defaultValue="overview" className="w-full p-4">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="info">Info</TabsTrigger>
        <TabsTrigger value="stats">Stats</TabsTrigger>
      </TabsList>
      <TabsContent value="overview">
        <div className="pb-4">
          <div className="font-semibold">Progress</div>
          <div className="flex flex-col items-center gap-1">
            <span className="text-muted-foreground text-center">
              {download.progress.toFixed(2)}%
            </span>
            <Progress value={download.progress} />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="font-semibold">Name</div>
            <span className="text-muted-foreground">{download.name}</span>
          </div>
          <div>
            <div className="font-semibold">Status</div>
            <span className="text-muted-foreground">
              {capitalizeFirstLetter(download.status)}
            </span>
          </div>
          <div>
            <div className="font-semibold">Address</div>
            <span className="text-muted-foreground">
              {capitalizeFirstLetter(download.url)}
            </span>
          </div>
          <div>
            <div className="font-semibold">Save Path</div>
            <span className="text-muted-foreground">
              {capitalizeFirstLetter(download.savePath)}
            </span>
          </div>
          <div>
            <div className="font-semibold">Added Date</div>
            <span className="text-muted-foreground">
              {new Date(download.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>
      </TabsContent>
      <TabsContent value="info">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="font-semibold">Name</div>
            <span className="text-muted-foreground">{download.name}</span>
          </div>
          <div>
            <div className="font-semibold">Total Size</div>
            <span className="text-muted-foreground">
              {(download.totalSize / 1024 / 1024).toFixed(2) + " MB"}
            </span>
          </div>
          <div>
            <div className="font-semibold">Added Date</div>
            <span className="text-muted-foreground">
              {new Date(download.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>
      </TabsContent>

      <TabsContent value="stats">
        <div className="pb-4">
          <div className="font-semibold">Progress</div>
          <div className="flex flex-col items-center gap-1">
            <span className="text-muted-foreground text-center">
              {download.progress.toFixed(2)}%
            </span>
            <Progress value={download.progress} />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="font-semibold">Status</div>
            <span className="text-muted-foreground">
              {capitalizeFirstLetter(download.status)}
            </span>
          </div>
          <div>
            <div className="font-semibold">Download Speed</div>
            <span className="text-muted-foreground">
              {download.downloadSpeed / 1024 + " MB/s"}
            </span>
          </div>
          <div>
            <div className="font-semibold">ETA</div>
            <span className="text-muted-foreground">{download.eta}</span>
          </div>
        </div>
      </TabsContent>
    </Tabs>
  )
}
