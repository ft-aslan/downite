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
import { Button } from "@/components/ui/button"
import { FileBrowserDialog } from "@/components/file-browser-dialog"
import { Dice1, Folder } from "lucide-react"
import React from "react"

interface DownloadFormProps {
  className?: string
  downloadMeta: components["schemas"]["DownloadMeta"]
  setOpen: (open: boolean) => void
}
export default function DownloadForm({
  className,
  downloadMeta,
  setOpen,
}: DownloadFormProps) {
  const [fileName, setFileName] = React.useState(downloadMeta.fileName)
  //TODO(fatih): dont use any as type. fegure out how we can type form for multipart form
  const form = useForm<components["schemas"]["DownloadReqBody"]>({
    defaultValues: {
      url: downloadMeta.url,
      savePath: "",
      startDownload: true,
      isIncompleteSavePathEnabled: false,
      incompleteSavePath: "",
      contentLayout: "Original",
      addTopOfQueue: false,
      category: "",
      overwrite: false,
      name: downloadMeta.fileName,
      tags: [],
    },
  })
  //TODO(fatih): dont use any as type. fegure out how we can type form for multipart form
  const onSubmit = form.handleSubmit((data) => {
    downloadFormMutation.mutate(data)
  })
  const watchIncompleteSavePath = form.watch(
    "isIncompleteSavePathEnabled",
    false
  )
  const downloadFormMutation = useMutation({
    mutationFn: async (data: components["schemas"]["DownloadReqBody"]) => {
      const res = await client.POST("/download", {
        body: data,
      })
      return res
    },
    onSuccess(result) {
      if (result.data) {
        toast("Download Started", {
          description: result.data.name,
        })
        form.reset()
        setOpen(false)
      }
    },
    onError(error) {
      toast.error(error.message)
    },
  })

  const getNewFileNameMutation = useMutation({
    mutationFn: async (
      data: components["schemas"]["GetNewFileNameForPathReqBody"]
    ) => {
      const res = await client.POST("/download/new-file-name", {
        body: data,
      })
      return res
    },
    onSuccess(result) {
      if (result.data) {
        setFileName(result.data)
      }
    },
    onError(error) {
      toast.error(error.message)
      setOpen(false)
    },
  })

  React.useEffect(() => {
    getNewFileNameMutation.mutate({
      fileName: downloadMeta.fileName,
      savePath: form.getValues("savePath"),
    })
  }, [])

  if (getNewFileNameMutation.isPending) {
    return null
  }

  return (
    <Form {...form}>
      <form
        onSubmit={onSubmit}
        className={cn("grid items-start gap-4", className)}
      >
        <div className="flex w-full flex-col gap-4 rounded-md border p-4">
          <div className="flex-1">
            <p className="mb-2 text-sm font-medium leading-none">Name: </p>
            <Input
              type="text"
              placeholder="Address"
              className="text-muted-foreground text-sm"
              value={downloadMeta.fileName}
              readOnly
            />
          </div>
          <div className="flex-1">
            <p className="mb-2 text-sm font-medium leading-none">Address: </p>
            <Input
              type="text"
              placeholder="Address"
              className="text-muted-foreground text-sm"
              value={downloadMeta.url}
              readOnly
            />
          </div>
          <div className="flex-1">
            <p className="mb-2 text-sm font-medium leading-none">Size: </p>
            <p className="text-muted-foreground text-sm">
              {(downloadMeta.totalSize / 1024 / 1024).toFixed(2) + " MB"}
            </p>
          </div>
        </div>
        <Card>
          <CardHeader>
            <CardDescription>Configure your download</CardDescription>
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
                      <Input type="text" placeholder="Save Path" {...field} />

                      <FileBrowserDialog
                        onSelect={(path) => {
                          field.onChange(path)
                          getNewFileNameMutation.mutate({
                            fileName: downloadMeta.fileName,
                            savePath: path,
                          })
                        }}
                      >
                        <Button variant="default" className="gap-1">
                          <Folder className="h-3.5 w-3.5" />
                          <span className="sm:whitespace-nowrap">Browse</span>
                        </Button>
                      </FileBrowserDialog>
                    </div>
                  </FormControl>
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
                    <FormLabel>Incomplete download path</FormLabel>
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
                            <span className="sm:whitespace-nowrap">Browse</span>
                          </Button>
                        </FileBrowserDialog>
                      </div>
                    </FormControl>
                    <FormDescription>
                      The path where the where download will be saved while it
                      is downloading
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
                        Use another path for incomplete download
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
              name="startDownload"
              render={({ field }) => (
                <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                  <div className="space-y-0.5">
                    <FormLabel className="text-base">Start download</FormLabel>
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
                    <Select value={field.value} onValueChange={field.onChange}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select download category" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="test">test</SelectItem>
                      </SelectContent>
                    </Select>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </CardContent>
        </Card>

        <LoadingButton type="submit" isLoading={downloadFormMutation.isPending}>
          Download
        </LoadingButton>
      </form>
    </Form>
  )
}
