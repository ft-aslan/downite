import createClient from "openapi-fetch"
import type { paths } from "./v1"

export const client = createClient<paths>({ baseUrl: "http://localhost:9999/" })
