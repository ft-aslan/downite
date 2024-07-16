import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer"
import { useState } from "react"
import { useMedia } from "react-use"
import { PlusCircle } from "lucide-react"

import { components } from "@/api/v1"
import GetDownloadMetaForm from "./GetDownloadMetaForm"
import DownloadForm from "./DownloadForm"

export function AddDownloadDialog({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(false)
  const [downloadMeta, setDownloadMeta] =
    useState<components["schemas"]["DownloadMeta"]>()
  const isDesktop = useMedia("(min-width: 768px)")

  const onOpenChange = (open: boolean) => {
    setDownloadMeta(undefined)
    setOpen(open)
  }
  if (isDesktop) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogTrigger asChild>{children}</DialogTrigger>
        {downloadMeta ? (
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Download</DialogTitle>
              <DialogDescription></DialogDescription>
            </DialogHeader>

            <DownloadForm downloadMeta={downloadMeta} setOpen={setOpen} />
          </DialogContent>
        ) : (
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Download</DialogTitle>
              <DialogDescription>
                You can add download with address.
              </DialogDescription>
            </DialogHeader>

            <GetDownloadMetaForm
              onDownloadMetaChange={(meta) => {
                setDownloadMeta(meta)
              }}
            />
          </DialogContent>
        )}
      </Dialog>
    )
  }

  return (
    <Drawer open={open} onOpenChange={setOpen}>
      <DrawerTrigger asChild>
        <Button variant="outline" className="gap-1">
          <PlusCircle className="h-3.5 w-3.5" />
          <span className="sm:whitespace-nowrap">Add Download</span>
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader className="text-left">
          <DrawerTitle>Add Download</DrawerTitle>
          <DrawerDescription>
            You can add download with address.
          </DrawerDescription>
        </DrawerHeader>
        <GetDownloadMetaForm onDownloadMetaChange={setDownloadMeta} />
        <DrawerFooter className="pt-2">
          <DrawerClose asChild>
            <Button variant="outline">Cancel</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}