import * as React from "react"
import {
  ColumnDef,
  ColumnFiltersState,
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

const columns: ColumnDef<components["schemas"]["Download"]>[] = [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={
          table.getIsAllPageRowsSelected() ||
          (table.getIsSomePageRowsSelected() && "indeterminate")
        }
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
        aria-label="Select all"
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
        aria-label="Select row"
      />
    ),
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
    cell: ({ row }) => <div>{row.getValue("queueNumber")}</div>,
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
    cell: ({ row }) => (
      <div>
        <Link to={`/download/$id`} params={{ id: row.original.id }}>
          {row.getValue("name")}
        </Link>
      </div>
    ),
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
    cell: ({ row }) => (
      <div className="flex flex-col items-center gap-2">
        <span>{(row.getValue("progress") as number).toFixed(2)}%</span>
        <Progress value={row.getValue("progress")} className="w-full" />
      </div>
    ),
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
    cell: ({ row }) => <div>{row.getValue("downloadSpeed")} KB/s</div>,
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
      const download = row.original

      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="h-8 w-8 p-0">
              <span className="sr-only">Open menu</span>
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuItem
              onClick={() => {
                client.POST("/download/pause", {
                  body: {
                    ids: [download.id],
                  },
                })
              }}
            >
              <Pause className="ml-2 h-4 w-4" />
              <span className="ml-1 sm:whitespace-nowrap">Pause</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => {
                client.POST("/download/resume", {
                  body: {
                    ids: [download.id],
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
                client.POST("/download/remove", {
                  body: {
                    ids: [download.id],
                  },
                })
              }}
            >
              <Trash2 className="ml-2 h-4 w-4" />
              <span className="ml-1 sm:whitespace-nowrap">Remove</span>
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => {
                client.POST("/download/delete", {
                  body: {
                    ids: [download.id],
                  },
                })
              }}
            >
              <Trash2 className="ml-2 h-4 w-4" />
              <span className="ml-1 sm:whitespace-nowrap">
                Delete With Files
              </span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      )
    },
  },
]

export function DownloadsTable({
  downloads,
}: {
  downloads: components["schemas"]["Download"][]
}) {
  const [view, setView] = React.useState<"table" | "grid">("table")
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  )
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({
      id: false,
    })
  const [rowSelection, setRowSelection] = React.useState({})

  const table = useReactTable({
    data: downloads,
    columns: columns,
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

  return (
    <div className="w-full">
      <div className="flex items-center py-4">
        <Input
          placeholder="Find Download..."
          value={(table.getColumn("name")?.getFilterValue() as string) ?? ""}
          onChange={(event) =>
            table.getColumn("name")?.setFilterValue(event.target.value)
          }
          className="max-w-sm"
        />
        <div className="ml-auto flex gap-2">
          <ToggleGroup
            type="single"
            value={view}
            onValueChange={(value: "table" | "grid") => setView(value)}
          >
            <ToggleGroupItem value="table" aria-label="Toggle table view">
              <TableIcon className="h-5 w-5" />
            </ToggleGroupItem>
            <ToggleGroupItem value="grid" aria-label="Toggle grid view">
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
      <div className="rounded-md border">
        {view === "table" ? (
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
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={columns.length}
                    className="h-24 text-center"
                  >
                    No results.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        ) : (
          <div>
            {table.getRowModel().rows?.map((row) => (
              <div key={row.id} className="relative flex rounded-sm px-2">
                <div
                  className={cn(
                    "absolute left-0 right-0 h-full w-1 rounded-l-sm",
                    row.getValue("state") === "downloading"
                      ? "bg-green-500"
                      : "bg-red-500"
                  )}
                ></div>
                {row.getVisibleCells().map((cell) => (
                  <div key={cell.id}>
                    <span className="text-muted-foreground"></span>
                    <span>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </span>
                  </div>
                ))}
              </div>
            ))}
          </div>
        )}
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
