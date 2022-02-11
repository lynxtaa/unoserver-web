import { createApp } from './app.js'

const PORT = Number(process.env.PORT) || 3000

const fastify = createApp({ basePath: process.env.BASE_PATH })

const ALL_AVAILABLE_IPV4_INTERFACES = '0.0.0.0'

await fastify.listen(PORT, ALL_AVAILABLE_IPV4_INTERFACES)

fastify.log.info(`Listening on port ${PORT}. Mode: ${process.env.NODE_ENV}`)
