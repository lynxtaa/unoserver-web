import { type AddressInfo } from 'node:net'

import { type FastifyInstance } from 'fastify'
import { Agent, fetch, type RequestInit, Response, setGlobalDispatcher } from 'undici'

import { createApp } from '../app.js'

export type TestServer = FastifyInstance & {
	fetch: (url: string, init?: RequestInit) => Promise<Response>
}

export async function startTestServer(): Promise<TestServer> {
	const app = createApp({ logLevel: 'warn' }) as unknown as TestServer

	await app.listen({ port: 0, host: 'localhost' })

	const { port } = app.server.address() as AddressInfo

	const agent = new Agent({ keepAliveTimeout: 10, keepAliveMaxTimeout: 10 })
	setGlobalDispatcher(agent)

	app.fetch = async (url, init) => {
		const response = await fetch(`http://localhost:${port}/${url}`, init)
		if (!response.ok) {
			throw new Error(`Bad response (${response.status})`)
		}
		return response
	}

	return app
}
