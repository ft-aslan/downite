import { components } from "@/api/v1"

import { Input } from "@/components/ui/input"
import { cn } from "@/lib/utils"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Label } from "@/components/ui/label"
import React from "react"
import { client } from "@/api"
import { toast } from "sonner"
import { useNavigate } from "@tanstack/react-router"
import { useMutation } from "@tanstack/react-query"

interface ExistingDownloadFormProps {
  className?: string
  downloadMeta: components["schemas"]["DownloadMeta"]
  setDownloadMeta: (value: components["schemas"]["DownloadMeta"]) => void
  setOpen: (open: boolean) => void
  setShowDownloadForm: (showDownloadForm: boolean) => void
}
export default function ExistingDownloadForm({
  className,
  downloadMeta,
  setOpen,
  setShowDownloadForm,
}: ExistingDownloadFormProps) {
  const [selectedOption, setSelectedOption] = React.useState("show")
  const navigate = useNavigate()

  const onRadioChange = (value: string) => {
    setSelectedOption(value)
  }

  const onSubmit = async () => {
    switch (selectedOption) {
      case "create":
        {
          setShowDownloadForm(true)
        }
        break
      case "resume":
        {
          let { error } = await client.POST("/download/resume", {
            body: { ids: [downloadMeta.existingDownloadId] },
          })
          if (error) {
            toast.error(`Failed to resume download: ${error.detail}`)
            return
          }
          toast.success(`Download ${downloadMeta.fileName} resumed`)
          setOpen(false)
        }
        break
      case "show":
        {
          navigate({ to: `/download/${downloadMeta.existingDownloadId}` })
          setOpen(false)
        }
        break
    }
  }
  return (
    <div className={cn("grid items-start gap-4", className)}>
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
          <CardDescription>
            Download is already exists in your downloads. You may choose one of
            following options.
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          <RadioGroup value={selectedOption} onValueChange={onRadioChange}>
            <div className="flex items-center gap-2">
              <RadioGroupItem className="w-auto" value="show" id="r4" />
              <Label className="text-md" htmlFor="r4">
                Show existing download
              </Label>
            </div>
            <div className="flex items-center gap-2">
              <RadioGroupItem className="w-auto" value="create" id="r2" />
              <Label className="text-md" htmlFor="r2">
                Create new duplicate with numbered file name
              </Label>
            </div>
            <div className="flex items-center gap-2">
              <RadioGroupItem className="w-auto" value="resume" id="r3" />
              <Label className="text-md text-wrap" htmlFor="r3">
                If existing one is incomplete and paused then resume download
              </Label>
            </div>
          </RadioGroup>
        </CardContent>
      </Card>

      <Button onClick={onSubmit}>Submit</Button>

      <Button variant="outline" onClick={() => setOpen(false)}>
        Cancel
      </Button>
    </div>
  )
}
