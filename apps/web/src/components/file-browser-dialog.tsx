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

export function FileBrowserDialog({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(false)
  const isDesktop = useMedia("(min-width: 768px)")

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
            <FileBrowser />
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

        <FileBrowser />
        <DrawerFooter className="pt-2">
          <DrawerClose asChild>
            <Button variant="outline">Cancel</Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  )
}
