import { Toaster } from "@/components/ui/sonner"
import {
  Outlet,
  createRootRoute,
  createRootRouteWithContext,
} from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/router-devtools"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"

import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import LeftNav from "./-components/LeftNav"

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient
}>()({
  component: RootComponent,
})
export default function RootComponent() {
  return (
    <>
      <div className="grid h-screen w-full pl-[56px]">
        <LeftNav />
        <div className="flex flex-col">
          <Outlet />
        </div>
      </div>
      <Toaster />
      {/* <ReactQueryDevtools initialIsOpen={false} /> */}
      <TanStackRouterDevtools position="bottom-right" />
    </>
  )
}
