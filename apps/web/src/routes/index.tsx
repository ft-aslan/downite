import { Button } from "@/components/ui/button"
import { Link, createFileRoute } from "@tanstack/react-router"
import { AddTorrentDialog } from "./-components/AddTorrentDialog"
import { PlusCircle } from "lucide-react"
import { TorrentsTable } from "./-components/TorrentsTable"

export const Route = createFileRoute("/")({
  component: HomePage,
})
export default function HomePage() {
  return <div>Welcome to Downite</div>
}
