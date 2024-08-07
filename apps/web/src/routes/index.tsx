import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/")({
  component: HomePage,
})

import { client } from "@/api"
import { components } from "@/api/v1"
import React from "react"
import { AreaChart } from "@/components/ui/area-chart"
// import { Area, AreaChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts"
export default function HomePage() {
  const [torrentSpeedsChartData, setTorrentSpeedsChartData] = React.useState<
    components["schemas"]["TorrentsTotalSpeedData"][]
  >([])

  const [downloadSpeedsChartData, setDownloadSpeedsChartData] = React.useState<
    components["schemas"]["DownloadsTotalSpeedData"][]
  >([])

  const fetchTorrentSpeeds = async () => {
    const { data: torrentSpeeds, error } = await client.GET("/torrent/speed")
    if (error) return
    if (!torrentSpeeds) return

    if (torrentSpeedsChartData.length > 100) {
      setTorrentSpeedsChartData([torrentSpeeds])
    } else {
      setTorrentSpeedsChartData((prevChartData) => [
        ...prevChartData,
        torrentSpeeds,
      ])
    }
  }
  const fetchDownloadSpeeds = async () => {
    const { data: downloadSpeeds, error } = await client.GET("/download/speed")
    if (error) return
    if (!downloadSpeeds) return

    if (torrentSpeedsChartData.length > 100) {
      setDownloadSpeedsChartData([downloadSpeeds])
    } else {
      setDownloadSpeedsChartData((prevChartData) => [
        ...prevChartData,
        downloadSpeeds,
      ])
    }
  }
  const fetchSpeeds = async () => {
    await fetchDownloadSpeeds()
    await fetchTorrentSpeeds()
  }
  React.useEffect(() => {
    fetchSpeeds()
    const tableUpdateInterval = setInterval(() => fetchSpeeds(), 1000)
    return () => clearInterval(tableUpdateInterval)
  }, [])

  return (
    <div className="p-4">
      <h1>Welcome to Downite</h1>
      <div className="mt-2 grid grid-cols-2 gap-4">
        <div>
          <header className="border-b pb-2">
            <h2 className="text-xl font-semibold">Torrents</h2>
          </header>
          <AreaChart
            className="h-80"
            data={torrentSpeedsChartData}
            index="time"
            categories={["downloadSpeed", "uploadSpeed"]}
          />
          {/* <AreaChart
            width={730}
            height={250}
            data={chartData}
            margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
          >
            <defs>
              <linearGradient id="colorUv" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
                <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
              </linearGradient>
              <linearGradient id="colorPv" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#82ca9d" stopOpacity={0.8} />
                <stop offset="95%" stopColor="#82ca9d" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis dataKey="time" />
            <YAxis />
            <CartesianGrid strokeDasharray="3 3" />
            <Tooltip />
            <Area
              type="monotone"
              dataKey="downloadSpeed"
              stroke="#8884d8"
              fillOpacity={1}
              fill="url(#colorUv)"
            />
            <Area
              type="monotone"
              dataKey="uploadSpeed"
              stroke="#82ca9d"
              fillOpacity={1}
              fill="url(#colorPv)"
            />
          </AreaChart>*/}
        </div>
        <div>
          <header className="border-b pb-2">
            <h2 className="text-xl font-semibold">Downloads</h2>
          </header>
          <AreaChart
            className="h-80"
            data={downloadSpeedsChartData}
            index="time"
            categories={["downloadSpeed"]}
          />
        </div>
      </div>
    </div>
  )
}
