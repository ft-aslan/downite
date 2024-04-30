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

const formSchema = z.object({
  magnet: z.string().optional(),
  torrentFile: z
    .instanceof(File)
    .refine(
      (file) => file.size > 0,
      "File size was 0, please upload a proper file!"
    )
    .optional(),
})
interface GetTorrentMetaFormProps {
  className?: string
  onTorrentMetaChange: (
    meta: components["schemas"]["TorrentMeta"],
    torrentFile?: File
  ) => void
}
export default function GetTorrentMetaForm({
  className,
  onTorrentMetaChange,
}: GetTorrentMetaFormProps) {
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      magnet: "",
      torrentFile: undefined,
    },
  })
  const torrentMetaFormMutation = useMutation({
    mutationFn: async (data: z.infer<typeof formSchema>) => {
      const byteArray = await data.torrentFile?.arrayBuffer()
      const array = byteArray && Array.from(new Uint8Array(byteArray))

      const res = await client.POST("/torrent-meta", {
        body: {
          ...data,
          torrentFile: array,
        },
      })
      return res
    },
    onSuccess(data) {
      if (data.data) {
        toast("Form Submitted", { description: JSON.stringify(data.data) })
        const { torrentFile: torrent } = form.getValues()
        form.reset()
        onTorrentMetaChange(data.data, torrent)
      }
    },
  })
  async function onSubmit(data: z.infer<typeof formSchema>) {
    if (data.torrentFile && data.magnet) {
      form.setError("magnet", {
        message: "There is both torrent file and magnet",
      })
      form.setError("torrentFile", {
        message: "There is both torrent file and magnet",
      })
      return
    }
    torrentMetaFormMutation.mutate(data)
  }
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("grid items-start gap-4", className)}
      >
        <FormField
          control={form.control}
          name="magnet"
          render={({ field }) => (
            <FormItem className="grid gap-2">
              <FormLabel>Magnet</FormLabel>
              <FormControl>
                <Input type="text" placeholder="magnet:?..." {...field} />
              </FormControl>
              <FormDescription>
                Enter Magnet Link to add Torrent
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="torrentFile"
          render={({ field }) => (
            <FormItem className="grid gap-2">
              <FormLabel>Torrent File</FormLabel>
              <FormControl>
                <Input
                  type="file"
                  placeholder="Torrent File"
                  accept=".torrent"
                  onChange={(e) => {
                    // only one file can be uploaded for now
                    if (e.target.files) {
                      field.onChange(e.target.files[0])
                    }
                  }}
                />
              </FormControl>
              <FormDescription>Upload Torrent File</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <LoadingButton
          type="submit"
          isLoading={torrentMetaFormMutation.isPending}
        >
          Get Metadata
        </LoadingButton>
      </form>
    </Form>
  )
}
