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
import FileBrowser from "./file-browser"
import React from "react"

export function FileBrowserDialog({
  children,
  onSelect,
}: {
  children: React.ReactNode
  onSelect: (path: string) => void
}) {
  const [open, setOpen] = useState(false)
  const isDesktop = useMedia("(min-width: 768px)")
  const [path, setPath] = React.useState("/")

  const onOpenChange = (open: boolean) => {
    setOpen(open)
  }
  if (isDesktop) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogTrigger asChild>{children}</DialogTrigger>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>File Browser</DialogTitle>
            <DialogDescription>Browse in file system</DialogDescription>
            <FileBrowser path={path} onChange={setPath} />
            <div>
              <Button
                className="w-full"
                onClick={() => {
                  onSelect(path)
                  setOpen(false)
                }}
              >
                Select
              </Button>
            </div>
          </DialogHeader>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Drawer open={open} onOpenChange={setOpen}>
      <DrawerTrigger asChild>{children}</DrawerTrigger>
      <DrawerContent>
        <DrawerHeader className="text-left">
          <DrawerTitle>File Browser</DrawerTitle>
          <DrawerDescription>Browse in file system</DrawerDescription>
        </DrawerHeader>

        <FileBrowser path={path} onChange={setPath} />
        <div>
          <Button className="w-full" onClick={() => {}}>
            Select
          </Button>
        </div>
        <DrawerFooter className="pt-2">
          <DrawerClose asChild>
            <Button variant="outline">Cancel</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}
