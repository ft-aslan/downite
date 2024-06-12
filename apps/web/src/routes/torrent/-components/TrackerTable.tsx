import { components } from "@/api/v1"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export default function TrackerTable({
  trackers,
}: {
  trackers: components["schemas"]["Tracker"][]
}) {
  return (
    <Table>
      <TableCaption>Torrent Trackers</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead>Tier</TableHead>
          <TableHead>URL</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Peers</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {trackers.length > 0 ? (
          trackers.map((tracker) => (
            <TableRow key={tracker.url}>
              <TableCell>{tracker.tier}</TableCell>
              <TableCell>{tracker.url}</TableCell>
              <TableCell></TableCell>
              <TableCell>{tracker.peers.length}</TableCell>
            </TableRow>
          ))
        ) : (
          <TableRow>
            <TableCell colSpan={3}>No trackers found</TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  )
}
