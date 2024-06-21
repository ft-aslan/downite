import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/torrent/$infohash")({
  component: () => TorrentRoot(),
  loader: ({ context: { queryClient }, params: { infohash } }) => {
    queryClient.ensureQueryData(getTorrentQueryOptions(infohash))
  },
})

import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { capitalizeFirstLetter } from "@/lib/utils"
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query"
import React from "react"
import { client } from "@/api"
import TorrentFileTree, {
  createFlatFileTree,
} from "@/components/TorrentFileTree"
import { Progress } from "@/components/ui/progress"
import PeerTable from "./-components/PeerTable"
import TrackerTable from "./-components/TrackerTable"

const getTorrentQueryOptions = (infohash: string) =>
  queryOptions({
    queryKey: ["torrent", infohash],
    queryFn: () =>
      client.GET("/torrent/{infohash}", {
        params: {
          path: {
            infohash,
          },
        },
      }),
  })
function TorrentRoot() {
  const { infohash } = Route.useParams()
  const {
    data: { data: torrent },
    refetch,
  } = useSuspenseQuery(getTorrentQueryOptions(infohash))
  const [fileTree, setFileTree] = React.useState(
    torrent?.files.map(createFlatFileTree) || []
  )
  React.useEffect(() => {
    const torrentUpdateInterval = setInterval(() => refetch(), 1000)
    return () => clearInterval(torrentUpdateInterval)
  }, [])

  if (!torrent) return null
  return (
    <Tabs defaultValue="overview">
      <TabsList>
        <TabsTrigger value="overview">Overview</TabsTrigger>
        <TabsTrigger value="info">Info</TabsTrigger>
        <TabsTrigger value="stats">Stats</TabsTrigger>
        <TabsTrigger value="peers">Peers</TabsTrigger>
        <TabsTrigger value="trackers">Trackers</TabsTrigger>
        <TabsTrigger value="files">Files</TabsTrigger>
      </TabsList>
      <TabsContent value="overview">
        <div className="pb-4">
          <div className="font-semibold">Progress</div>
          <div className="flex flex-col items-center gap-1">
            <span className="text-muted-foreground text-center">
              {torrent.progress.toFixed(2)}%
            </span>
            <Progress value={torrent.progress} />
          </div>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="font-semibold">Name</div>
            <span className="text-muted-foreground">{torrent.name}</span>
          </div>
          <div>
            <div className="font-semibold">Status</div>
            <span className="text-muted-foreground">
              {capitalizeFirstLetter(torrent.status)}
            </span>
          </div>

          <div>
            <div className="font-semibold">Size</div>
            <span className="text-muted-foreground">
              {(torrent.sizeOfWanted / 1024 / 1024).toFixed(2) + " MB"}
            </span>
          </div>
          <div>
            <div className="font-semibold">Added Date</div>
            <span className="text-muted-foreground">
              {new Date(torrent.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>
      </TabsContent>
      <TabsContent value="info">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="font-semibold">Name</div>
            <span className="text-muted-foreground">{torrent.name}</span>
          </div>
          <div>
            <div className="font-semibold">Total Size</div>
            <span className="text-muted-foreground">
              {(torrent.totalSize / 1024 / 1024).toFixed(2) + " MB"}
            </span>
          </div>
          <div>
            <div className="font-semibold">Size</div>
            <span className="text-muted-foreground">
              {(torrent.sizeOfWanted / 1024 / 1024).toFixed(2) + " MB"}
            </span>
          </div>
          <div>
            <div className="font-semibold">Added Date</div>
            <span className="text-muted-foreground">
              {new Date(torrent.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>
      </TabsContent>

      <TabsContent value="stats">
        <div className="pb-4">
          <div className="font-semibold">Progress</div>
          <div className="flex flex-col items-center gap-1">
            <span className="text-muted-foreground text-center">
              {torrent.progress.toFixed(2)}%
            </span>
            <Progress value={torrent.progress} />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="font-semibold">Status</div>
            <span className="text-muted-foreground">
              {capitalizeFirstLetter(torrent.status)}
            </span>
          </div>
          <div>
            <div className="font-semibold">Download Speed</div>
            <span className="text-muted-foreground">
              {torrent.downloadSpeed / 1024 + " MB/s"}
            </span>
          </div>
          <div>
            <div className="font-semibold">Upload Speed</div>
            <span className="text-muted-foreground">
              {torrent.uploadSpeed / 1024 + " MB/s"}
            </span>
          </div>
          <div>
            <div className="font-semibold">Peers</div>
            <span className="text-muted-foreground">{torrent.peerCount}</span>
          </div>
          <div>
            <div className="font-semibold">Seeds</div>
            <span className="text-muted-foreground">{torrent.seeds}</span>
          </div>
          <div>
            <div className="font-semibold">Ratio</div>
            <span className="text-muted-foreground">{torrent.ratio}</span>
          </div>
          <div>
            <div className="font-semibold">ETA</div>
            <span className="text-muted-foreground">{torrent.eta}</span>
          </div>
        </div>
      </TabsContent>
      <TabsContent value="peers">
        <PeerTable peers={torrent.peers} />
      </TabsContent>
      <TabsContent value="trackers">
        <TrackerTable trackers={torrent.trackers} />
      </TabsContent>
      <TabsContent value="files">
        <ScrollArea className="h-80">
          <TorrentFileTree fileTree={fileTree} setFileTree={setFileTree} />
        </ScrollArea>
      </TabsContent>
    </Tabs>
  )
}
