import assert from 'node:assert/strict'
import { createReadStream } from 'node:fs'
import { rm, stat } from 'node:fs/promises'
import path from 'node:path'

import contentDisposition from 'content-disposition'
import { type FastifyPluginCallback } from 'fastify'
import httpErrors from 'http-errors'
import mime from 'mime-types'

import { convertFile } from './utils/convertFile.js'
import { upload } from './utils/upload.js'

export const routes: FastifyPluginCallback = (app, options, next) => {
	app.post<{ Params: { format: string }; Querystring: { filter: string } }>(
		'/convert/:format',
		{
			preHandler: upload.single('file'),
			schema: {
				summary: 'Converts file using LibreOffice',
				consumes: ['multipart/form-data'],
				produces: ['application/octet-stream'],
				params: { format: { type: 'string' } },
				querystring: { filter: { type: 'string' } },
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
			assert(req.file !== undefined, new httpErrors.BadRequest('Expected file'))

			const { path: srcPath, destination } = req.file

			assert(
				srcPath !== undefined && destination !== undefined,
				'Expected "path" and "destination"',
			)

			res.raw.on('close', () => {
				rm(destination, { recursive: true }).catch(() => {
					// ignore
				})
			})

			const { targetPath } = await convertFile(srcPath, req.params.format, {
				filter: req.query.filter,
			})

			const stream = createReadStream(targetPath)

			const mimeType = mime.lookup(req.params.format)

			res.type(mimeType === false ? 'application/octet-stream' : mimeType)
			res.header('Content-Disposition', contentDisposition(path.parse(targetPath).base))

			const { size } = await stat(targetPath)
			res.header('Content-Length', size)

			res.send(stream)

			return res
		},
	)

	next()
}
