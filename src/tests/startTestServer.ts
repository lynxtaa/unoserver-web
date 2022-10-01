import { AddressInfo } from 'node:net'

import { FastifyInstance } from 'fastify'
import { Agent, request, setGlobalDispatcher } from 'undici'

import { createApp } from '../app.js'

export type TestServer = FastifyInstance & {
	request: (
		url: string,
		init?: Parameters<typeof request>[1],
	) => ReturnType<typeof request>
}

export async function startTestServer(): Promise<TestServer> {
	const app = createApp({ logLevel: 'warn' }) as unknown as TestServer

	await app.listen({ port: 0, host: 'localhost' })

	const { port } = app.server.address() as AddressInfo

	const agent = new Agent({ keepAliveTimeout: 10, keepAliveMaxTimeout: 10 })
	setGlobalDispatcher(agent)

	app.request = async (url, init) => {
		const response = await request(`http://localhost:${port}/${url}`, {
			throwOnError: true,
			...init,
		})
		return response
	}

	return app
}
