import { Toaster } from "@/components/ui/sonner"
import { Outlet, createRootRoute } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/router-devtools"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"

import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import LeftNav from "./-components/LeftNav"

const queryClient = new QueryClient()

export const Route = createRootRoute({
  component: RootComponent,
})
export default function RootComponent() {
  return (
    <QueryClientProvider client={queryClient}>
      <div className="grid h-screen w-full pl-[56px]">
        <LeftNav />
        <div className="flex flex-col">
          <Outlet />
        </div>
      </div>
      <Toaster />
      {/* <ReactQueryDevtools initialIsOpen={false} /> */}
      <TanStackRouterDevtools position="bottom-right" />
    </QueryClientProvider>
  )
}
