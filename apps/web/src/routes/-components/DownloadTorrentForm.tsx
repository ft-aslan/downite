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
import { Tree } from "@/components/ui/tree-view"
import { File as FileIcon, Folder, Layout } from "lucide-react"
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

const formSchema = z.object({
  magnet: z.string().startsWith("magnet:?").optional(),
  torrentFile: z.instanceof(File).optional(),
  savePath: z.string(),
  isIncompleteSavePathEnabled: z.boolean().default(false),
  incompleteSavePath: z.string().optional(),
  category: z.string().optional(),
  tags: z.string().array().optional(),
  startTorrent: z.boolean().default(true),
  addTopOfQueue: z.boolean().default(false),
  downloadSequentially: z.boolean().default(false),
  skipHashCheck: z.boolean().default(false),
  contentLayout: z.string().default("Original"),
  files: z
    .object({
      path: z.string(),
      name: z.string(),
      wanted: z.boolean().default(true),
      downloadPriority: z.string().default("Normal"),
    })
    .array(),
})
interface GetTorrentMetaFormProps {
  className?: string
  torrentMeta?: components["schemas"]["TorrentMeta"]
}
export default function DownloadTorrentForm({
  className,
  torrentMeta,
}: GetTorrentMetaFormProps) {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      startTorrent: true,
    },
  })
  const watchIncompleteSavePath = form.watch(
    "isIncompleteSavePathEnabled",
    false
  )
  const torrentDownloadFormMutation = useMutation({
    mutationFn: async (data: z.infer<typeof formSchema>) => {
      const res = await client.POST("/torrent", {
        body: data,
      })
      return res
    },
    onSuccess(data) {
      if (data.data) {
        toast("Form Submitted", { description: JSON.stringify(data.data) })
        form.reset()
      }
    },
  })
  async function onSubmit(data: z.infer<typeof formSchema>) {
    torrentDownloadFormMutation.mutate(data)
  }
  const data = [
    { id: "1", name: "Unread" },
    { id: "2", name: "Threads" },
    {
      id: "6",
      name: "Direct Messages",
      children: [
        {
          id: "f1",
          name: "Alice",
          children: [
            { id: "f11", name: "Alice2" },
            { id: "f12", name: "Bob2" },
            { id: "f13", name: "Charlie2" },
          ],
        },
        { id: "f2", name: "Bob" },
        { id: "f3", name: "Charlie" },
      ],
    },
  ]
  const [tab, setTab] = React.useState("config")
  const onTabChange = (value: string) => {
    setTab(value)
  }
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("grid items-start gap-4", className)}
      >
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
                <Tree
                  className="h-64 w-full"
                  data={data}
                  onSelectChange={(item) => console.log(item)}
                  folderIcon={Folder}
                  itemIcon={FileIcon}
                />
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
