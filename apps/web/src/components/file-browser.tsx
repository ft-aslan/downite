import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useQuery } from "@tanstack/react-query"
import { client } from "@/api"
import { ArrowBigLeftIcon, FileIcon, FolderIcon, Loader2 } from "lucide-react"
import { ScrollArea } from "@/components/ui/scroll-area"
import { cn } from "@/lib/utils"

export default function FileBrowser({
  path,
  onChange,
}: {
  path: string
  onChange: (path: string) => void
}) {
  const { data, error, isLoading, refetch } = useQuery({
    queryKey: ["nodes", path],
    queryFn: () => {
      return client.POST("/os/filesystem", {
        body: {
          path,
        },
      })
    },
  })
  if (error) {
    return <div>Error</div>
  }
  return (
    <div className="space-y-4">
      <Input value={path} onChange={(e) => onChange(e.target.value)} />
      <ScrollArea className="h-[500px] rounded-md border p-4">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-4">Type</TableHead>
              <TableHead>Name</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.data?.fileSystemNodes.map((node) => (
              <TableRow
                key={node.path}
                onClick={() => {
                  if (node.type === "file") {
                    return
                  }
                  onChange(node.path)
                  refetch()
                }}
              >
                <TableCell>
                  {node.type === "file" ? (
                    <FileIcon className="text-muted h-4 w-4" />
                  ) : node.type === "dir" ? (
                    <FolderIcon className="h-4 w-4" />
                  ) : (
                    <ArrowBigLeftIcon className="h-4 w-4" />
                  )}
                </TableCell>
                <TableCell
                  className={cn(
                    "select-none font-medium",
                    node.type === "file"
                      ? "text-muted-foreground"
                      : "text-foreground cursor-pointer"
                  )}
                >
                  {node.name}
                </TableCell>
                <TableCell className="w-4">
                  {isLoading && node.path === path ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </ScrollArea>
    </div>
  )
}
