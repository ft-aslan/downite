import React from "react"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useQuery } from "@tanstack/react-query"
import { client } from "@/api"
import { FileIcon, FolderIcon } from "lucide-react"

export default function FileBrowser() {
  const [path, setPath] = React.useState("/")
  const { data, error } = useQuery({
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
    <div>
      <Input value={path} onChange={(e) => setPath(e.target.value)} />
      <Table>
        <TableCaption>Files</TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead>Type</TableHead>
            <TableHead>Name</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data?.data?.fileSystemNodes.map((node) => (
            <TableRow key={node.path}>
              <TableCell>
                {node.type === "file" ? <FileIcon /> : <FolderIcon />}
              </TableCell>
              <TableCell>{node.path}</TableCell>
            </TableRow>
          ))}
        </TableBody>
        <TableFooter>
          <TableRow>
            <TableCell colSpan={2}>{path}</TableCell>
          </TableRow>
        </TableFooter>
      </Table>
    </div>
  )
}
