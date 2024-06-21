import { Toaster } from "@/components/ui/sonner"
import { Outlet, createRootRouteWithContext } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/router-devtools"

import { QueryClient } from "@tanstack/react-query"
import LeftNav from "./-components/LeftNav"
import { client } from "@/api"
import { CircleX } from "lucide-react"

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient
}>()({
  component: RootComponent,
  loader: async () => {
    try {
      const res = await client.GET("/torrent")
      return res
    } catch (error) {
      return null
    }
  },
})
export default function RootComponent() {
  const res = Route.useLoaderData()
  return (
    <>
      <div className="grid h-screen w-full pl-[56px]">
        <LeftNav />
        <div className="flex flex-col">
          {!res && (
            <div className="w-full bg-red-700 p-2 text-center text-white">
              <CircleX className="m-auto h-8 w-8" />
              <p className="font-bold">Server is unreachable !!!</p>
            </div>
          )}
          <Outlet />
        </div>
      </div>
      <Toaster />
      {/* <ReactQueryDevtools initialIsOpen={false} /> */}
      <TanStackRouterDevtools position="bottom-right" />
    </>
  )
}
