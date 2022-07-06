import { createApp } from './app.js'

const PORT = process.env.PORT ? Number(process.env.PORT) : 3000

const fastify = createApp({ basePath: process.env.BASE_PATH })

const ALL_AVAILABLE_IPV4_INTERFACES = '0.0.0.0'

await fastify.listen({ port: PORT, host: ALL_AVAILABLE_IPV4_INTERFACES })

fastify.log.info(`Mode: ${process.env.NODE_ENV}`)
