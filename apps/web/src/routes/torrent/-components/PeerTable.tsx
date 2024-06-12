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

export default function PeerTable({
  peers,
}: {
  peers: components["schemas"]["Peer"][]
}) {
  return (
    <Table>
      <TableCaption>Torrent Peers</TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead>URL</TableHead>
          <TableHead>Port</TableHead>
          <TableHead>Connection</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {peers.length > 0 ? (
          peers.map((peer) => (
            <TableRow key={peer.url}>
              <TableCell>{peer.url}</TableCell>
              <TableCell></TableCell>
              <TableCell></TableCell>
            </TableRow>
          ))
        ) : (
          <TableRow>
            <TableCell colSpan={3}>No peers found</TableCell>
          </TableRow>
        )}
      </TableBody>
    </Table>
  )
}
