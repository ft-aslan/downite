import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/")({
  component: HomePage,
})

import { client } from "@/api"
import { components } from "@/api/v1"
import { queryOptions, useSuspenseQuery } from "@tanstack/react-query"
import React from "react"
import { AreaChart } from "@/components/ui/area-chart"
export default function HomePage() {
  const [chartData, setChartData] = React.useState<
    components["schemas"]["TorrentsTotalSpeedData"][]
  >([])

  React.useEffect(() => {
    const tableUpdateInterval = setInterval(async () => {
      const { data, error } = await client.GET("/torrent/speed")
      if (error) return
      if (!data) return

      if (chartData.length > 100) {
        setChartData([data])
      } else {
        setChartData([...chartData, data])
      }
    }, 1000)
    return () => clearInterval(tableUpdateInterval)
  }, [])

  return (
    <div>
      <div>Welcome to Downite</div>
      <AreaChart
        className="h-80"
        data={chartData}
        index="time"
        categories={["downloadSpeed", "uploadSpeed"]}
      />
    </div>
  )
}
