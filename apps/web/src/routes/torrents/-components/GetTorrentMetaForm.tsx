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

interface GetTorrentMetaFormProps {
  type: "magnet" | "file"
  onTorrentMetaChange: (
    meta: components["schemas"]["TorrentMeta"],
    torrentFile?: File
  ) => void
}
export default function GetTorrentMetaForm({
  type,
  onTorrentMetaChange,
}: GetTorrentMetaFormProps) {
  if (type === "magnet") {
    return <WithMagnet onTorrentMetaChange={onTorrentMetaChange} />
  } else {
    return <WithFile onTorrentMetaChange={onTorrentMetaChange} />
  }
}
interface WithFileProps {
  onTorrentMetaChange: (
    meta: components["schemas"]["TorrentMeta"],
    torrentFile?: File
  ) => void
}
function WithFile({ onTorrentMetaChange }: WithFileProps) {
  const form = useForm({})
  const torrentMetaFormMutation = useMutation({
    mutationFn: async (data: MultipartFormData) => {
      const res = await client.POST("/meta/file", {
        body: data,
        bodySerializer(body) {
          //turn it into multipart/form-data by bypassing json serialization
          const fd = new FormData()
          for (const name in body) {
            const field = body[name]
            if (field) {
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
        toast("Form Submitted", { description: JSON.stringify(result.data) })
        const { torrentFile: torrent } = form.getValues()
        form.reset()
        onTorrentMetaChange(result.data, torrent)
      }
    },
  })
  async function onSubmit(data) {
    torrentMetaFormMutation.mutate(data)
  }
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("grid items-start gap-4")}
      >
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
                    if (e.target.files?.length === 1) {
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
interface WithMagnetProps {
  onTorrentMetaChange: (
    meta: components["schemas"]["TorrentMeta"],
    torrentFile?: File
  ) => void
}
function WithMagnet({ onTorrentMetaChange }: WithMagnetProps) {
  const form = useForm<components["schemas"]["GetMetaWithMagnetReqBody"]>({
    defaultValues: {
      magnet: "",
    },
  })
  const torrentMetaFormMutation = useMutation({
    mutationFn: async (
      data: components["schemas"]["GetMetaWithMagnetReqBody"]
    ) => {
      const res = await client.POST("/meta/magnet", {
        body: data,
      })
      return res
    },
    onSuccess(result) {
      if (result.data) {
        toast("Form Submitted", { description: JSON.stringify(result.data) })
        form.reset()
        onTorrentMetaChange(result.data, undefined)
      }
    },
  })
  async function onSubmit(
    data: components["schemas"]["GetMetaWithMagnetReqBody"]
  ) {
    torrentMetaFormMutation.mutate(data)
  }
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("grid items-start gap-4")}
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
