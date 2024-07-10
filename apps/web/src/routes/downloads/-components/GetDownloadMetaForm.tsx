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

interface GetDownloadMetaFormProps {
  onDownloadMetaChange: (meta: components["schemas"]["DownloadMeta"]) => void
}
export default function GetDownloadMetaForm({
  onDownloadMetaChange,
}: GetDownloadMetaFormProps) {
  const form = useForm<components["schemas"]["GetDownloadMetaReqBody"]>({
    defaultValues: {
      url: "",
    },
  })
  const downloadMetaFormMutation = useMutation({
    mutationFn: async (
      data: components["schemas"]["GetDownloadMetaReqBody"]
    ) => {
      const res = await client.POST("/download/meta", {
        body: data,
      })
      return res
    },
    onSuccess(result) {
      if (result.data) {
        toast(
          `Fetched download metadata from address : ${result.data.fileName}`
        )
        form.reset()
        onDownloadMetaChange(result.data)
      }
    },
  })
  async function onSubmit(
    data: components["schemas"]["GetDownloadMetaReqBody"]
  ) {
    downloadMetaFormMutation.mutate(data)
  }
  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn("grid items-start gap-4")}
      >
        <FormField
          control={form.control}
          name="url"
          render={({ field }) => (
            <FormItem className="grid gap-2">
              <FormLabel>Address</FormLabel>
              <FormControl>
                <Input type="text" placeholder="Enter Address..." {...field} />
              </FormControl>
              <FormDescription>Enter Address</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <LoadingButton
          type="submit"
          isLoading={downloadMetaFormMutation.isPending}
        >
          Get Metadata
        </LoadingButton>
      </form>
    </Form>
  )
}
