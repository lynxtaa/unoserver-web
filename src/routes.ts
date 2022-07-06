import { createReadStream } from 'fs'
import { rm, stat } from 'fs/promises'
import path from 'path'

import contentDisposition from 'content-disposition'
import { FastifyPluginCallback } from 'fastify'
import httpErrors from 'http-errors'
import mime from 'mime-types'

import { assert } from './utils/assert.js'
import { convertFile } from './utils/convertFile.js'
import { upload } from './utils/upload.js'

export const routes: FastifyPluginCallback = (app, options, next) => {
	app.post<{ Params: { format: string } }>(
		'/convert/:format',
		{
			preHandler: upload.single('file'),
			schema: {
				summary: 'Converts file using LibreOffice',
				consumes: ['multipart/form-data'],
				produces: ['application/octet-stream'],
				params: { format: { type: 'string' } },
				body: {
					properties: { file: { type: 'string', format: 'binary' } },
					required: ['file'],
				},
				response: {
					'200': {},
				},
			},
		},
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

			const { size } = await stat(targetPath)
			res.header('Content-Length', size)

			res.send(stream)

			return res
		},
	)

	next()
}
