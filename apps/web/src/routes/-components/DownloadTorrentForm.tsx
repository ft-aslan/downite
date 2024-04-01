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

const formSchema = z.object({
  magnet: z.string().startsWith("magnet:?").optional(),
  torrent: z.instanceof(File).optional(),
  savePath: z.string(),
  category: z.string(),
  tags: z.string().array(),
  startTorrent: z.boolean().default(true),
  addTopOfQueue: z.boolean().default(false),
  downloadSequentially: z.boolean().default(false),
  skipHashCheck: z.boolean().default(false),
  contentLayout: z.string().default("Original"),
  files: z.object({
    path: z.string(),
    name: z.string(),
    wanted: z.boolean().default(true),
    downloadPriority: z.string().default("Normal"),
  }),
})
interface GetTorrentMetaFormProps {
  className?: string
  torrentMeta?: components["schemas"]["GetTorrentMetaRes"]
}
export default function DownloadTorrentForm({
  className,
  torrentMeta,
}: GetTorrentMetaFormProps) {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {},
  })
  const torrentDownloadFormMutation = useMutation({
    mutationFn: async (data: z.infer<typeof formSchema>) => {
      const res = await client.POST("/api/v1/torrent-meta", {
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
  const [content, setContent] = React.useState("Admin Page")
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("grid items-start gap-4", className)}
      >
        <FormField
          control={form.control}
          name="savePath"
          render={({ field }) => (
            <FormItem className="grid gap-2">
              <FormLabel>Save At</FormLabel>
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

        <Tree
          className="h-64 w-full"
          data={data}
          onSelectChange={(item) => setContent(item?.name ?? "")}
          folderIcon={Folder}
          itemIcon={FileIcon}
        />

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
