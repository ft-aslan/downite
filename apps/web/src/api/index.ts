import createClient from "openapi-fetch"
import useWebSocket from 'react-use-websocket';
import type { paths } from "./v1"

const domain = "localhost:9999";

export const client = createClient<paths>({
  baseUrl: `http://${domain}/api`,
})


export function useSocketClient<T>(path: string) {
  return useWebSocket<T>(`ws://${domain}/api${path}`);
}