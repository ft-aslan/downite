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
// import { z } from "zod"

// import { zodResolver } from "@hookform/resolvers/zod"
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
import { ScrollArea } from "@/components/ui/scroll-area"
import TorrentFileTree, {
  createFlatFileTree,
} from "@/components/TorrentFileTree"
import { FileBrowserDialog } from "@/components/file-browser-dialog"
import { Button } from "@/components/ui/button"
import { Folder } from "lucide-react"

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
  const [fileTree, setFileTree] = React.useState(
    torrentMeta.files.map(createFlatFileTree)
  )
  // components["schemas"]["DownloadTorrentWithMagnetReqBody"]

  //TODO(fatih): dont use any as type. fegure out how we can type form for multipart form
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
  //TODO(fatih): dont use any as type. fegure out how we can type form for multipart form
  const onSubmit = form.handleSubmit((data) => {
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
      data.torrentFile = torrentFile
    } else {
      data.magnet = torrentMeta.magnet
    }

    torrentDownloadFormMutation.mutate(data)
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
        bodySerializer(body) {
          //turn it into multipart/form-data by bypassing json serialization
          const fd = new FormData()
          for (const name in body) {
            const field = (body as any)[name]
            if (Array.isArray(field)) {
              fd.append(name, JSON.stringify(field))
            } else {
              fd.append(name, field)
            }
          }
          return fd
        },
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
        onSubmit={onSubmit}
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
                        <div className="flex flex-row items-center justify-between gap-2 rounded-lg border p-4">
                          <Input
                            type="text"
                            placeholder="Save Path"
                            {...field}
                          />

                          <FileBrowserDialog
                            onSelect={(path) => field.onChange(path)}
                          >
                            <Button variant="default" className="gap-1">
                              <Folder className="h-3.5 w-3.5" />
                              <span className="sm:whitespace-nowrap">
                                Browse
                              </span>
                            </Button>
                          </FileBrowserDialog>
                        </div>
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
                          <div className="flex flex-row items-center justify-between gap-2 rounded-lg border p-4">
                            <Input
                              type="text"
                              placeholder="Incomplete Save Path"
                              {...field}
                            />

                            <FileBrowserDialog
                              onSelect={(path) => field.onChange(path)}
                            >
                              <Button variant="default" className="gap-1">
                                <Folder className="h-3.5 w-3.5" />
                                <span className="sm:whitespace-nowrap">
                                  Browse
                                </span>
                              </Button>
                            </FileBrowserDialog>
                          </div>
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
                <ScrollArea className="h-80">
                  <TorrentFileTree
                    fileTree={fileTree}
                    setFileTree={setFileTree}
                  />
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
}
