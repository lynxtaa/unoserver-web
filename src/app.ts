import { type Server, type IncomingMessage, type ServerResponse } from 'node:http'

import cors from '@fastify/cors'
import swagger from '@fastify/swagger'
import swaggerUi from '@fastify/swagger-ui'
import Fastify, { type FastifyInstance } from 'fastify'
import multer from 'fastify-multer'
import { type P, pino } from 'pino'

import { transform } from './plugins/supportFilesInSchema.js'
import { routes } from './routes.js'
import { unoserver } from './utils/unoserver.js'

// https://github.com/fox1t/fastify-multer/blob/master/typings/fastify/index.d.ts
interface File {
	fieldname: string
	originalname: string
	encoding: string
	mimetype: string
	size?: number
	destination?: string
	filename?: string
	path?: string
	buffer?: Buffer
	stream?: NodeJS.ReadableStream
}

declare module 'fastify' {
	interface FastifyRequest {
		file: File
		files: Record<string, File[]> | Partial<File>[]
	}
}

export function createApp({
	basePath = '',
	logLevel,
	requestIdHeader,
	requestIdLogLabel,
}: {
	basePath?: string
	logLevel?: P.LevelWithSilent
	requestIdHeader?: string
	requestIdLogLabel?: string
} = {}): FastifyInstance<Server, IncomingMessage, ServerResponse, P.Logger> {
	const fastify = Fastify({
		trustProxy: true,
		requestIdHeader,
		requestIdLogLabel,
		logger: pino({
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
		}),
	})

	fastify.register(cors, { origin: '*', maxAge: 60 * 60 })

	fastify.register(multer.contentParser)

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
