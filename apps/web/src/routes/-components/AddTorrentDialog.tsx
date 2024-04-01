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
import { useEffect, useState } from "react"
import { useMedia } from "react-use"
import { PlusCircle } from "lucide-react"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card"
import { components } from "@/api/v1"
import GetTorrentMetaForm from "./GetTorrentMetaFrom"
import DownloadTorrentForm from "./DownloadTorrentForm"

export function AddTorrentRename() {
  const [open, setOpen] = useState(false)
  const [torrentMeta, setTorrentMeta] =
    useState<components["schemas"]["GetTorrentMetaRes"]>()

  const [tab, setTab] = useState("get-torrent-meta")
  const onTabChange = (value: string) => {
    setTab(value)
  }

  useEffect(() => {
    if (torrentMeta) {
      // if we have torrent meta, we can show the download tab
      setTab("download")
    }
  }, [torrentMeta])

  const isDesktop = useMedia("(min-width: 768px)")

  if (isDesktop) {
    return (
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogTrigger asChild>
          <Button variant="default" className="gap-1">
            <PlusCircle className="h-3.5 w-3.5" />
            <span className="sm:whitespace-nowrap">Add Torrent</span>
          </Button>
        </DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Torrent</DialogTitle>
            <DialogDescription></DialogDescription>
          </DialogHeader>
          <Tabs
            value={tab}
            onValueChange={onTabChange}
            defaultValue="get-torrent-meta"
          >
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="get-torrent-meta">Send Meta</TabsTrigger>
              <TabsTrigger
                value="download"
                disabled={torrentMeta ? false : true}
              >
                Download
              </TabsTrigger>
            </TabsList>
            <TabsContent value="get-torrent-meta">
              <Card>
                <CardHeader>
                  <CardDescription>
                    You can add torrents with magnet or torrent file.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <GetTorrentMetaForm onTorrentMetaChange={setTorrentMeta} />
                </CardContent>
              </Card>
            </TabsContent>
            <TabsContent value="download">
              <Card>
                <CardHeader>
                  <CardDescription>Configure torrent download</CardDescription>
                </CardHeader>
                <CardContent>
                  <DownloadTorrentForm onTorrentMetaChange={setTorrentMeta} />
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </DialogContent>
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
        <GetTorrentMetaForm
          className="px-4"
          onTorrentMetaChange={setTorrentMeta}
        />
        <DrawerFooter className="pt-2">
          <DrawerClose asChild>
            <Button variant="outline">Cancel</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}
