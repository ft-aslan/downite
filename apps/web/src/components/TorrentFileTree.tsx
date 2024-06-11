import { components } from "@/api/v1"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { ChevronRight, FileIcon, Folder } from "lucide-react"
type DownloadPriority = components["schemas"]["TorrentFileTreeNode"]["priority"]

export interface FileTreeNode {
  id: string
  name: string
  size: string
  path: string
  priority: DownloadPriority
  expanded: boolean
  children: FileTreeNode[]
}
export const createFlatFileTree = (
  file: components["schemas"]["TorrentFileTreeNode"]
): FileTreeNode => ({
  id: file.path,
  name: file.name,
  size: (file.length / 1024 / 1024).toFixed(2) + " MB",
  path: file.path,
  priority: "normal",
  expanded: false,
  children: file.children.map(createFlatFileTree),
})
export default function TorrentFileTree({
  fileTree,
  setFileTree,
}: {
  fileTree: FileTreeNode[]
  setFileTree: React.Dispatch<React.SetStateAction<FileTreeNode[]>>
}) {
  const updateTreeNodeById = (
    id: string,
    updateFn: (item: FileTreeNode) => FileTreeNode
  ): void => {
    const updateRecursive = (obj: FileTreeNode): FileTreeNode => {
      if (obj.id === id) {
        return updateFn(obj)
      }
      if (obj.children) {
        return { ...obj, children: obj.children.map(updateRecursive) }
      }
      return obj
    }

    setFileTree((prevData) => prevData.map(updateRecursive))
  }
  return (
    <Table>
      <TableCaption>Torrent file tree</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Size</TableHead>
          <TableHead>Download Priority</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {fileTree.map((item) => (
          <RenderFileTree key={item.id} item={item} level={0} />
        ))}
      </TableBody>
    </Table>
  )

  function RenderFileTree({
    item,
    level,
  }: {
    item: FileTreeNode
    level: number
  }) {
    return [
      <TableRow key={item.id}>
        <TableCell>
          <div className="flex items-center space-x-2">
            <div className="h-2" style={{ width: `${level * 45}px` }}></div>
            {item.children.length > 0 && (
              <Button
                variant="ghost"
                size="sm"
                className="w-9"
                style={{ transform: `rotate(${item.expanded ? 90 : 0}deg)` }}
                onClick={() => {
                  updateTreeNodeById(item.id, (item) => ({
                    ...item,
                    expanded: !item.expanded,
                  }))
                }}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            )}
            <Checkbox
              checked={item.priority != "none"}
              onCheckedChange={(checked) => {
                updateTreeNodeById(item.id, (item) => {
                  // if changed to folder checkbox, set download priority for all children recursively
                  if (item.children.length) {
                    item.children.forEach((child) => {
                      child.priority = checked ? "normal" : "none"
                    })
                  }
                  return {
                    ...item,
                    priority: checked ? "normal" : "none",
                  }
                })
              }}
            />
            {item.children.length ? (
              <Folder className="h-4 w-4" />
            ) : (
              <FileIcon className="h-4 w-4" />
            )}
            <span className="min-w-8">{item.name}</span>
          </div>
        </TableCell>
        <TableCell>
          <span className="whitespace-nowrap">{item.size}</span>
        </TableCell>
        <TableCell>
          <Select
            value={item.priority}
            onValueChange={(downloadPriority: DownloadPriority) => {
              updateTreeNodeById(item.id, (item) => ({
                ...item,
                priority: downloadPriority,
              }))
            }}
          >
            <SelectTrigger>
              <SelectValue placeholder="Select download priority" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={"none"}>None</SelectItem>
              <SelectItem value={"low"}>Low</SelectItem>
              <SelectItem value={"normal"}>Normal</SelectItem>
              <SelectItem value={"high"}>High</SelectItem>
              <SelectItem value={"maximum"}>Maximum</SelectItem>
            </SelectContent>
          </Select>
        </TableCell>
      </TableRow>,
      ...(item.expanded
        ? item.children.map((child) => (
            <RenderFileTree key={child.id} item={child} level={level + 1} />
          ))
        : []),
    ]
  }
}
