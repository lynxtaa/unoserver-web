import { AddressInfo } from 'net'

import { FastifyInstance } from 'fastify'
import fetch, { RequestInit, Response } from 'node-fetch'

import { createApp } from '../app.js'
import { assert } from '../utils/assert.js'

export type TestServer = FastifyInstance & {
	fetch: (url: string, init?: RequestInit) => Promise<Response>
}

export async function startTestServer(): Promise<TestServer> {
	const app = createApp({ logLevel: 'warn' }) as unknown as TestServer

	await app.listen(0, 'localhost')

	const { port } = app.server.address() as AddressInfo

	app.fetch = async (url, init) => {
		const response = await fetch(`http://localhost:${port}/${url}`, init)
		assert(response.ok, `Error response ${response.status} from ${response.url}`)
		return response
	}

	return app
}
