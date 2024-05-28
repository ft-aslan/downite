import { client } from "@/api"
import { components } from "@/api/v1"
import {
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage,
  Form,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { LoadingButton } from "@/components/ui/loading-button"
import { cn } from "@/lib/utils"
import { useMutation } from "@tanstack/react-query"
import { useForm } from "react-hook-form"
import { toast } from "sonner"
import { z } from "zod"

import { zodResolver } from "@hookform/resolvers/zod"
import { ChevronRight, File as FileIcon, Folder } from "lucide-react"
import React from "react"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card"
import { Switch } from "@/components/ui/switch"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Checkbox } from "@/components/ui/checkbox"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Button } from "@/components/ui/button"

// const formSchema = z.object({
//   magnet: z.string().startsWith("magnet:?").optional(),
//   torrentFile: z.instanceof(File).optional(),
//   savePath: z.string(),
//   isIncompleteSavePathEnabled: z.boolean().default(false),
//   incompleteSavePath: z.string().optional(),
//   category: z.string().optional(),
//   tags: z.string().array().optional(),
//   startTorrent: z.boolean().default(true),
//   addTopOfQueue: z.boolean().default(false),
//   downloadSequentially: z.boolean().default(false),
//   skipHashCheck: z.boolean().default(false),
//   contentLayout: z.string().default("Original"),
//   files: z
//     .object({
//       name: z.string(),
//       path: z.string(),
//       downloadPriority: z.string().default("Normal"),
//     })
//     .array(),
// })
interface GetTorrentMetaFormProps {
  className?: string
  torrentMeta: components["schemas"]["TorrentMeta"]
  torrentFile?: File
  setOpen: (open: boolean) => void
}
export default function DownloadTorrentForm({
  className,
  torrentMeta,
  torrentFile,
  setOpen,
}: GetTorrentMetaFormProps) {
  const form = useForm<components["schemas"]["DownloadTorrentReqBody"]>({
    defaultValues: {
      savePath: "",
      startTorrent: true,
      files: [],
      isIncompleteSavePathEnabled: false,
      contentLayout: "Original",
      addTopOfQueue: false,
      downloadSequentially: false,
      skipHashCheck: false,
      category: "",
      tags: [],
    },
  })
  const watchIncompleteSavePath = form.watch(
    "isIncompleteSavePathEnabled",
    false
  )
  const torrentDownloadFormMutation = useMutation({
    mutationFn: async (
      data: components["schemas"]["DownloadTorrentReqBody"]
    ) => {
      const res = await client.POST("/torrent", {
        body: data,
      })
      return res
    },
    onSuccess(result) {
      if (result.data) {
        toast("Torrent Download Started", {
          description: result.data.name,
        })
        form.reset()
        setOpen(false)
      }
    },
  })
  async function onSubmit(
    data: components["schemas"]["DownloadTorrentReqBody"]
  ) {
    data.files = fileTree.map((file) => ({
      name: file.name,
      path: file.path,
      priority: file.priority,
    }))
    if (!torrentMeta.magnet && !torrentFile) {
      // TODO(fatih): Show error to user. There is no torrent file or magnet
      return
    }

    if (torrentFile) {
      data.torrentFile = await torrentFile.text()
    } else {
      data.magnet = torrentMeta.magnet
    }

    torrentDownloadFormMutation.mutate(data)
  }
  type DownloadPriority =
    components["schemas"]["TorrentFileTreeNode"]["priority"]

  interface FileTreeNode {
    id: string
    name: string
    size: string
    path: string
    priority: DownloadPriority
    expanded: boolean
    children: FileTreeNode[]
  }
  const createFileTree = (
    file: components["schemas"]["TorrentFileTreeNode"]
  ): FileTreeNode => ({
    id: file.path,
    name: file.name,
    size: (file.length / 1024 / 1024).toFixed(2) + " MB",
    path: file.path,
    priority: "normal",
    expanded: false,
    children: file.children.map(createFileTree),
  })
  const [fileTree, setFileTree] = React.useState(
    torrentMeta.files.map(createFileTree)
  )
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
  const [tab, setTab] = React.useState("config")
  const onTabChange = (value: string) => {
    setTab(value)
  }
  if (!torrentMeta.magnet && !torrentFile) {
    return (
      <div>
        <p>There is no torrent file or magnet</p>
      </div>
    )
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("grid items-start gap-4", className)}
      >
        <div className="flex flex-col space-y-2 rounded-md border p-4">
          <div className="flex-1 space-y-1">
            <p className="text-sm font-medium leading-none">Name:</p>
            <p className="text-muted-foreground text-sm">{torrentMeta.name}</p>
          </div>
          <div className="flex-1 space-y-1">
            <p className="text-sm font-medium leading-none">Info Hash: </p>
            <p className="text-muted-foreground text-sm">
              {torrentMeta.infohash}
            </p>
          </div>
          <div className="flex-1 space-y-1">
            <p className="text-sm font-medium leading-none">Size: </p>
            <p className="text-muted-foreground text-sm">
              {(torrentMeta.totalSize / 1024 / 1024).toFixed(2) + " MB"}
            </p>
          </div>
        </div>
        <Tabs
          value={tab}
          onValueChange={onTabChange}
          defaultValue="get-torrent-meta"
        >
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="config">Config</TabsTrigger>
            <TabsTrigger value="fileTree">File Tree</TabsTrigger>
          </TabsList>
          <TabsContent value="config">
            <Card>
              <CardHeader>
                <CardDescription>
                  Configure your torrent download
                </CardDescription>
              </CardHeader>
              <CardContent className="grid gap-4">
                <FormField
                  control={form.control}
                  name="savePath"
                  render={({ field }) => (
                    <FormItem className="grid gap-2">
                      <FormLabel>Save path</FormLabel>
                      <FormControl>
                        <Input type="text" placeholder="Save Path" {...field} />
                      </FormControl>
                      <FormDescription>
                        The path where the torrent will be saved
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {watchIncompleteSavePath && (
                  <FormField
                    control={form.control}
                    name="incompleteSavePath"
                    render={({ field }) => (
                      <FormItem className="grid gap-2">
                        <FormLabel>Incomplete torrent path</FormLabel>
                        <FormControl>
                          <Input
                            type="text"
                            placeholder="Incomplete Save Path"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription>
                          The path where the where torrent will be saved while
                          it is downloading
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}

                <FormField
                  control={form.control}
                  name="isIncompleteSavePathEnabled"
                  render={({ field }) => (
                    <FormItem className="grid gap-4">
                      <div className="flex flex-row items-center justify-between rounded-lg border p-4">
                        <div className="space-y-0.5">
                          <FormLabel className="text-base">
                            Use another path for incomplete torrent
                          </FormLabel>
                          <FormDescription>
                            Use another path for incomplete torrent
                          </FormDescription>
                        </div>
                        <FormControl>
                          <Switch
                            checked={field.value}
                            onCheckedChange={field.onChange}
                          />
                        </FormControl>
                      </div>

                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="startTorrent"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">
                          Start torrent
                        </FormLabel>
                        <FormDescription>
                          Start the torrent after creation
                        </FormDescription>
                      </div>

                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>

                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="addTopOfQueue"
                  render={({ field }) => (
                    <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                      <div className="space-y-0.5">
                        <FormLabel className="text-base">
                          Add top of queue
                        </FormLabel>
                        <FormDescription>
                          Add the torrent to the top of the queue
                        </FormDescription>
                      </div>

                      <FormControl>
                        <Switch
                          checked={field.value}
                          onCheckedChange={field.onChange}
                        />
                      </FormControl>

                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="category"
                  render={({ field }) => (
                    <FormItem className="grid gap-2">
                      <FormLabel>Category</FormLabel>
                      <FormControl>
                        <Select
                          value={field.value}
                          onValueChange={field.onChange}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue placeholder="Select torrent category" />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="m@example.com">
                              m@example.com
                            </SelectItem>
                            <SelectItem value="m@google.com">
                              m@google.com
                            </SelectItem>
                            <SelectItem value="m@support.com">
                              m@support.com
                            </SelectItem>
                          </SelectContent>
                        </Select>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="fileTree">
            <Card>
              <CardHeader>
                <CardDescription>Torrent file tree editor</CardDescription>
              </CardHeader>
              <CardContent>
                {/* <Tree
                  className="h-64 w-full"
                  data={fileTree}
                  onSelectChange={(item) => console.log(item)}
                  folderIcon={Folder}
                  itemIcon={FileIcon}
                /> */}
                <ScrollArea className="h-80">
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
                </ScrollArea>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        <LoadingButton
          type="submit"
          isLoading={torrentDownloadFormMutation.isPending}
        >
          Download Torrent
        </LoadingButton>
      </form>
    </Form>
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
