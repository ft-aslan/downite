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
  magnet: z.string().startsWith("magnet:?").optional(),
  torrent: z.instanceof(File).optional(),
})
interface GetTorrentMetaFormProps {
  className?: string
  onTorrentMetaChange: (
    data: components["schemas"]["GetTorrentMetaRes"]
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
      torrent: undefined,
    },
  })
  const torrentMetaFormMutation = useMutation({
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
        onTorrentMetaChange(data.data)
      }
    },
  })
  async function onSubmit(data: z.infer<typeof formSchema>) {
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
          name="torrent"
          render={({ field }) => (
            <FormItem className="grid gap-2">
              <FormLabel>Torrent</FormLabel>
              <FormControl>
                <Input
                  type="file"
                  placeholder="Torrent File"
                  {...field}
                  accept=".torrent"
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
