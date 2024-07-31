import * as React from "react"
import { atom, useAtom } from "jotai"
import {
  ColumnDef,
  ColumnFiltersState,
  RowSelectionState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table"
import {
  ArrowUpDown,
  Check,
  ChevronDown,
  MoreHorizontal,
  Pause,
  Play,
  RefreshCw,
  Rows3,
  Table as TableIcon,
  Trash2,
} from "lucide-react"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { components } from "@/api/v1"
import { client } from "@/api"
import { Progress } from "@/components/ui/progress"
import { Link } from "@tanstack/react-router"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { cn } from "@/lib/utils"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { toast } from "sonner"
import { DefaultAlertDialog } from "@/components/default-alert-dialog"
function DownloadStatusIcon(props: { status: string }) {
  switch (props.status) {
    case "metadata":
      return <RefreshCw className="h-4 w-4" />
    case "paused":
      return <Pause className="h-4 w-4" />
    case "downloading":
      return <Play className="h-4 w-4" />
    case "completed":
      return <Check className="h-4 w-4" />
    default:
      return null
  }
}
function toggleDownloadState(id: number, status: string) {
  if (status === "downloading") {
    client.POST("/download/pause", {
      body: {
        ids: [id],
      },
    })
  } else if (status === "paused") {
    client.POST("/download/resume", {
      body: {
        ids: [id],
      },
    })
  }
}

const columns = (
  view: "table" | "list",
  isCheckboxVisible: boolean
): ColumnDef<components["schemas"]["Download"]>[] => {
  return [
    {
      id: "select",
      header: ({ table }) => {
        if (view === "list" && !isCheckboxVisible) {
          return null
        }
        return (
          <Checkbox
            checked={
              table.getIsAllPageRowsSelected() ||
              (table.getIsSomePageRowsSelected() && "indeterminate")
            }
            onCheckedChange={(value) =>
              table.toggleAllPageRowsSelected(!!value)
            }
            aria-label="Select all"
          />
        )
      },
      cell: ({ row }) => {
        if (view === "list" && !isCheckboxVisible) {
          return null
        }
        return (
          <Checkbox
            checked={row.getIsSelected()}
            onCheckedChange={(value) => row.toggleSelected(!!value)}
            aria-label="Select row"
          />
        )
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "queueNumber",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            #
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => {
        if (view === "list") {
          return (
            <div className="flex flex-col">
              <span className="text-muted-foreground">#</span>
              <span>{row.getValue("queueNumber")}</span>
            </div>
          )
        }
        return <div>{row.getValue("queueNumber")}</div>
      },
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => (
        <Button
          variant="ghost"
          onClick={() =>
            toggleDownloadState(row.getValue("id"), row.getValue("status"))
          }
        >
          <DownloadStatusIcon status={row.getValue("status")} />
        </Button>
      ),
    },
    {
      accessorKey: "name",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Name
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => {
        return (
          <Link to={`/download/$id`} params={{ id: row.original.id }}>
            {row.getValue("name")}
          </Link>
        )
      },
    },
    {
      accessorKey: "progress",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Progress
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => {
        return (
          <div
            className={cn(
              "flex",
              "flex-col",
              "items-center",
              "gap-2",
              view === "list" ? "w-1/4" : ""
            )}
          >
            <span>{(row.getValue("progress") as number).toFixed(2)}%</span>
            <Progress value={row.getValue("progress")} className="w-full" />
          </div>
        )
      },
    },
    {
      accessorKey: "downloadSpeed",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Download Speed
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => {
        if (view === "list") {
          return (
            <div className="flex flex-col">
              <span className="text-muted-foreground">Download Speed</span>
              <span>{row.getValue("downloadSpeed")} KB/s</span>
            </div>
          )
        }
        return <div>{row.getValue("downloadSpeed")} KB/s</div>
      },
    },
    {
      accessorKey: "downloadedBytes",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Downloaded
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => {
        if (view === "list") {
          return (
            <div className="flex flex-col">
              <span className="text-muted-foreground">Downloaded</span>
              <span>
                {(
                  (row.getValue("downloadedBytes") as number) /
                  1024 /
                  1024
                ).toFixed(2)}{" "}
                MB
              </span>
            </div>
          )
        }
        return (
          <div>
            {(
              (row.getValue("downloadedBytes") as number) /
              1024 /
              1024
            ).toFixed(2)}{" "}
            MB
          </div>
        )
      },
    },
    {
      accessorKey: "totalSize",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            Total Size
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => {
        if (view === "list") {
          return (
            <div className="flex flex-col">
              <span className="text-muted-foreground">Total Size</span>
              <span>
                {((row.getValue("totalSize") as number) / 1024 / 1024).toFixed(
                  2
                )}{" "}
                MB
              </span>
            </div>
          )
        }
        return (
          <div>
            {((row.getValue("totalSize") as number) / 1024 / 1024).toFixed(2)}{" "}
            MB
          </div>
        )
      },
    },
    {
      accessorKey: "isMultiPart",
      header: "Resumable",
      cell: ({ row }) => {
        if (view === "list") {
          return (
            <div className="flex flex-col">
              <span className="text-muted-foreground">Resumable</span>

              {row.getValue("isMultiPart") ? (
                <span>Yes</span>
              ) : (
                <span className="text-red-500">No</span>
              )}
            </div>
          )
        }
        return row.getValue("isMultiPart") ? "Yes" : "No"
      },
    },
    {
      accessorKey: "id",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          >
            ID
            <ArrowUpDown className="ml-2 h-4 w-4" />
          </Button>
        )
      },
      cell: ({ row }) => <div>{row.getValue("id")}</div>,
    },
    {
      id: "actions",
      enableHiding: false,
      cell: ({ row }) => {
        if (isCheckboxVisible) {
          return null
        }
        const download = row.original

        return (
          <div className="flex flex-1 justify-end">
            <DownloadActionsDropdown downloadIds={[download.id]}>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DownloadActionsDropdown>
          </div>
        )
      },
    },
  ]
}
function DownloadActionsDropdown({
  downloadIds,
  children,
}: {
  downloadIds: number[]
  children?: React.ReactNode
}) {
  const isSingleDownload = downloadIds.length === 1
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>{children}</DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>Actions</DropdownMenuLabel>
        <DropdownMenuItem
          onClick={() => {
            if (downloadIds.length === 0) {
              toast.error("No downloads selected")
              return
            }
            client.POST("/download/pause", {
              body: {
                ids: downloadIds,
              },
            })
          }}
        >
          <Pause className="ml-2 h-4 w-4" />
          <span className="ml-1 sm:whitespace-nowrap">Pause</span>
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => {
            if (downloadIds.length === 0) {
              toast.error("No downloads selected")
              return
            }
            client.POST("/download/resume", {
              body: {
                ids: downloadIds,
              },
            })
          }}
        >
          <Play className="ml-2 h-4 w-4" />
          <span className="ml-1 sm:whitespace-nowrap">Resume</span>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            if (downloadIds.length === 0) {
              toast.error("No downloads selected")
              return
            }
            client.POST("/download/remove", {
              body: {
                ids: downloadIds,
              },
            })
          }}
        >
          <Trash2 className="ml-2 h-4 w-4" />
          <span className="ml-1 sm:whitespace-nowrap">Remove</span>
        </DropdownMenuItem>
        <DefaultAlertDialog
          title="Delete With Files"
          description={
            isSingleDownload
              ? "Are you sure you want to delete this download?"
              : "Are you sure you want to delete these downloads?"
          }
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={() => {
            if (downloadIds.length === 0) {
              toast.error("No downloads selected")
              return
            }
            client.POST("/download/delete", {
              body: {
                ids: downloadIds,
              },
            })
          }}
        >
          <DropdownMenuItem>
            <Trash2 className="ml-2 h-4 w-4" />
            <span className="ml-1 sm:whitespace-nowrap">Delete With Files</span>
          </DropdownMenuItem>
        </DefaultAlertDialog>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
export function DownloadsTable({
  downloads,
}: {
  downloads: components["schemas"]["Download"][]
}) {
  const [view, setView] = React.useState<"table" | "list">(
    (localStorage.getItem("view") as "table" | "list") ?? "table"
  )
  const [isSelectModeEnabled, setIsSelectModeEnabled] = React.useState(false)
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  )
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      id: false,
    })
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({})

  const table = useReactTable({
    data: downloads,
    columns: columns(view, isSelectModeEnabled),
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  })
  const TableView = () => (
    <Table>
      <TableHeader>
        {table.getHeaderGroups().map((headerGroup) => (
          <TableRow key={headerGroup.id}>
            {headerGroup.headers.map((header) => {
              return (
                <TableHead key={header.id}>
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                </TableHead>
              )
            })}
          </TableRow>
        ))}
      </TableHeader>
      <TableBody>
        {table.getRowModel().rows?.length ? (
          table.getRowModel().rows.map((row) => (
            <TableRow
              key={row.id}
              data-state={row.getIsSelected() && "selected"}
            >
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))
        ) : (
          <TableRow>
            <TableCell colSpan={columns.length} className="h-24 text-center">
              No results.
            </TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  )
  const ListView = () => (
    <div className="grid grid-cols-1 gap-4">
      {table.getRowModel().rows?.map((row) => {
        const nameCell = row
          .getVisibleCells()
          .find((cell) => cell.column.id === "name")
        const statusCell = row
          .getVisibleCells()
          .find((cell) => cell.column.id === "status")
        return (
          <div key={row.id} className="relative rounded-md border px-2">
            <div
              className={cn(
                "absolute left-0 right-0 h-full w-1 rounded-l-sm",
                row.getValue("status") === "downloading"
                  ? "bg-primary"
                  : row.getValue("status") === "error"
                    ? "bg-red-500"
                    : "bg-gray-700"
              )}
            ></div>
            <div className="flex place-items-center">
              <div className="p-2">
                {statusCell != undefined
                  ? flexRender(
                      statusCell.column.columnDef.cell,
                      statusCell.getContext()
                    )
                  : null}
              </div>
              <div className="grid w-full p-2">
                <div className="w-full text-ellipsis">
                  {nameCell != undefined
                    ? flexRender(
                        nameCell.column.columnDef.cell,
                        nameCell.getContext()
                      )
                    : null}
                </div>
                <div className="flex flex-wrap items-center gap-4">
                  {row
                    .getVisibleCells()
                    .filter(
                      (cell) =>
                        cell.column.id !== "name" && cell.column.id !== "status"
                    )
                    .map((cell) => (
                      <React.Fragment key={cell.id}>
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </React.Fragment>
                    ))}
                </div>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
  return (
    <div className="w-full">
      <div className="space-y-2 py-2">
        <div>
          <Input
            placeholder="Find Download..."
            value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
            onChange={(event) =>
              table.getColumn("name")?.setFilterValue(event.target.value)
            }
            className="max-w-sm"
          />
        </div>
        <div className="flex items-center">
          {view === "list" ? (
            <div className="flex gap-2">
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Checkbox
                      checked={isSelectModeEnabled}
                      onCheckedChange={(value) => setIsSelectModeEnabled(value)}
                      className="mr-2"
                    />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Toggle Select Mode</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
          ) : (
            <></>
          )}
          <div className="ml-auto flex gap-2">
            <ToggleGroup
              type="single"
              value={view}
              onValueChange={(value: "table" | "list") => {
                setView(value)
                localStorage.setItem("view", value)
              }}
            >
              <ToggleGroupItem value="table" aria-label="Toggle table view">
                <TableIcon className="h-5 w-5" />
              </ToggleGroupItem>
              <ToggleGroupItem value="list" aria-label="Toggle grid view">
                <Rows3 className="h-5 w-5" />
              </ToggleGroupItem>
            </ToggleGroup>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline">
                  Columns <ChevronDown className="ml-2 h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {table
                  .getAllColumns()
                  .filter((column) => column.getCanHide())
                  .map((column) => {
                    return (
                      <DropdownMenuCheckboxItem
                        key={column.id}
                        className="capitalize"
                        checked={column.getIsVisible()}
                        onCheckedChange={(value) =>
                          column.toggleVisibility(!!value)
                        }
                      >
                        {column.id}
                      </DropdownMenuCheckboxItem>
                    )
                  })}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        {view === "list" && isSelectModeEnabled ? (
          <div className="flex items-center gap-2">
            <Checkbox
              checked={
                table.getIsAllPageRowsSelected() ||
                (table.getIsSomePageRowsSelected() && "indeterminate")
              }
              onCheckedChange={(value) =>
                table.toggleAllPageRowsSelected(!!value)
              }
              aria-label="Select all"
            />
            <div className="flex items-center">
              <DownloadActionsDropdown
                downloadIds={Object.keys(rowSelection).flatMap((key) => {
                  if (rowSelection[key]) {
                    return downloads[key].id
                  }
                })}
              >
                <Button variant="outline">
                  <MoreHorizontal className="mr-2 h-4 w-4" />
                  <span>Actions</span>
                </Button>
              </DownloadActionsDropdown>
            </div>
          </div>
        ) : (
          <></>
        )}
      </div>
      <div className="rounded-md">
        {view === "table" ? <TableView /> : <ListView />}
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="text-muted-foreground flex-1 text-sm">
          {table.getFilteredSelectedRowModel().rows.length} of{" "}
          {table.getFilteredRowModel().rows.length} row(s) selected.
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  )
}
