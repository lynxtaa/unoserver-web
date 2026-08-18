import assert from 'node:assert/strict'
import { createReadStream } from 'node:fs'
import { rm, stat } from 'node:fs/promises'
import path from 'node:path'

import { create as createCD } from 'content-disposition'
import { type FastifyPluginCallback } from 'fastify'
import httpErrors from 'http-errors'
import mime from 'mime-types'

import { convertFile } from './utils/convertFile.js'
import { createRandomFolder } from './utils/upload.js'

export const routes: FastifyPluginCallback = (app, options, next) => {
	app.post<{ Params: { format: string }; Querystring: { filter: string } }>(
		'/convert/:format',
		{
			schema: {
				summary: 'Converts file using LibreOffice',
				consumes: ['multipart/form-data'],
				produces: ['application/octet-stream'],
				params: {
					type: 'object',
					properties: { format: { type: 'string' } },
				},
				querystring: {
					type: 'object',
					properties: { filter: { type: 'string' } },
				},
				body: {
					type: 'object',
					properties: { file: { type: 'string', format: 'binary' } },
					required: ['file'],
				},
				response: {
					'200': {},
				},
			},
			// Trickery to satisfy schema validation
			preValidation: (request, reply, done) => {
				request.body = { file: '' }
				done()
			},
		},
		async (req, res) => {
			const randomFolder = await createRandomFolder()

			res.raw.on('close', () => {
				rm(randomFolder, { recursive: true }).catch(() => {
					// ignore
				})
			})

			const { files } = await req.saveRequestFiles({
				tmpdir: randomFolder,
			})

			assert(files[0], new httpErrors.BadRequest('Expected file'))

			const [{ filepath: srcPath, filename }] = files

			const { targetPath } = await convertFile(srcPath, req.params.format, {
				filter: req.query.filter,
			})

			const stream = createReadStream(targetPath)

			const mimeType = mime.lookup(req.params.format)

			res.type(mimeType === false ? 'application/octet-stream' : mimeType)
			res.header(
				'Content-Disposition',
				createCD(path.parse(filename).name + path.parse(targetPath).ext),
			)

			const { size } = await stat(targetPath)
			res.header('Content-Length', size)

			res.send(stream)

			return res
		},
	)

	next()
}
