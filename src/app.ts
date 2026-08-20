import { type Server, type IncomingMessage, type ServerResponse } from 'node:http'

import cors from '@fastify/cors'
import multipart from '@fastify/multipart'
import swagger from '@fastify/swagger'
import swaggerUi from '@fastify/swagger-ui'
import Fastify, { type FastifyBaseLogger, type FastifyInstance } from 'fastify'
import { type LevelWithSilent } from 'pino'

import { transform } from './plugins/supportFilesInSchema.js'
import { routes } from './routes.js'
import { unoserver } from './utils/unoserver.js'

export function createApp({
	basePath = '',
	logLevel,
	requestIdHeader,
	requestIdLogLabel,
	maxFileSize = Infinity,
}: {
	basePath?: string
	logLevel?: LevelWithSilent
	requestIdHeader?: string
	requestIdLogLabel?: string
	maxFileSize?: number
} = {}): FastifyInstance<Server, IncomingMessage, ServerResponse, FastifyBaseLogger> {
	const fastify = Fastify({
		trustProxy: true,
		requestIdHeader,
		requestIdLogLabel,
		logger: {
			base: null,
			timestamp: false,
			level: logLevel ?? (process.env.NODE_ENV === 'production' ? 'info' : 'debug'),
			transport:
				process.env.NODE_ENV !== 'production'
					? {
							target: 'pino-pretty',
							options: { colorize: true },
						}
					: undefined,
		},
	})

	fastify.register(cors, { origin: '*', maxAge: 60 * 60 })

	// Without an explicit limit @fastify/multipart falls back to Fastify's bodyLimit (1 MiB)
	fastify.register(multipart, { limits: { fileSize: maxFileSize } })

	fastify.register(swagger, {
		swagger: {
			basePath: basePath === '' ? undefined : basePath,
			info: {
				title: 'unoserver-web',
				version: '0.1.0',
			},
			consumes: ['application/json'],
			produces: ['application/json'],
		},
		transform,
	})

	fastify.register(swaggerUi, {
		routePrefix: '/documentation',
		initOAuth: {},
	})

	fastify.get('/', (req, res) => {
		res.redirect(`${basePath}/documentation/static/index.html`)
	})

	fastify.register(routes)

	fastify.addHook('onRequest', (request, reply, done) => {
		request.raw.setTimeout(0)
		done()
	})

	fastify.addHook('onClose', () => {
		unoserver.stopServer()
	})

	return fastify
}
