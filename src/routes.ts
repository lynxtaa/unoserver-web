import { createReadStream, readFileSync } from 'fs'
import { rm } from 'fs/promises'
import path from 'path'

import contentDisposition from 'content-disposition'
import { FastifyPluginCallback } from 'fastify'
import httpErrors from 'http-errors'
import mime from 'mime-types'

import { assert } from './utils/assert.js'
import { convertFile } from './utils/convertFile.js'
import { upload } from './utils/upload.js'

const schema = JSON.parse(readFileSync('./routes.schema.json', 'utf-8'))

export const routes: FastifyPluginCallback = (app, options, next) => {
	app.post<{ Params: { format: string } }>(
		'/convert/:format',
		{ preHandler: upload.single('file'), schema: schema['/convert/:format'] },
		async (req, res) => {
			assert(req.file, new httpErrors.BadRequest('Expected file'))

			const { path: srcPath, destination } = req.file

			assert(srcPath && destination, 'Expected "path" and "destination"')

			res.raw.on('close', () => {
				rm(destination, { recursive: true }).catch(() => {
					// ignore
				})
			})

			const { targetPath } = await convertFile(srcPath, req.params.format)

			const stream = createReadStream(targetPath)

			res.type(mime.lookup(req.params.format) || 'application/octet-stream')
			res.header('Content-Disposition', contentDisposition(path.parse(targetPath).base))

			res.send(stream)
		},
	)

	next()
}
