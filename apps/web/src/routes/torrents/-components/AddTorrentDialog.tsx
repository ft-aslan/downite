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
import GetTorrentMetaForm from "./GetTorrentMetaForm"
import DownloadTorrentForm from "./DownloadTorrentForm"

export function AddTorrentDialog({
  type,
  children,
}: {
  type: "magnet" | "file"
  children: React.ReactNode
}) {
  const [open, setOpen] = useState(false)
  const [torrentMeta, setTorrentMeta] =
    useState<components["schemas"]["TorrentMeta"]>()
  const [torrentFile, setTorrentFile] = useState<File>()
  const isDesktop = useMedia("(min-width: 768px)")

  const onOpenChange = (open: boolean) => {
    setTorrentMeta(undefined)
    setTorrentFile(undefined)
    setOpen(open)
  }
  if (isDesktop) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogTrigger asChild>{children}</DialogTrigger>
        {torrentMeta ? (
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Torrent</DialogTitle>
              <DialogDescription></DialogDescription>
            </DialogHeader>

            <DownloadTorrentForm
              torrentMeta={torrentMeta}
              torrentFile={torrentFile}
              setOpen={setOpen}
            />
          </DialogContent>
        ) : (
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Torrent</DialogTitle>
              <DialogDescription>
                {`You can add torrents with ${type === "magnet" ? "magnet link" : "torrent file"}.`}
              </DialogDescription>
            </DialogHeader>

            <GetTorrentMetaForm
              type={type}
              onTorrentMetaChange={(meta, torrentFile) => {
                setTorrentMeta(meta)
                setTorrentFile(torrentFile)
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
          <span className="sm:whitespace-nowrap">Add Torrent</span>
        </Button>
      </DrawerTrigger>
      <DrawerContent>
        <DrawerHeader className="text-left">
          <DrawerTitle>Add Torrent</DrawerTitle>
          <DrawerDescription>
            You can add torrents with magnet or torrent file.
          </DrawerDescription>
        </DrawerHeader>
        <GetTorrentMetaForm type={type} onTorrentMetaChange={setTorrentMeta} />
        <DrawerFooter className="pt-2">
          <DrawerClose asChild>
            <Button variant="outline">Cancel</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}
