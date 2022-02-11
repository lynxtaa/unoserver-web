import { mkdir } from 'fs/promises'
import { tmpdir } from 'os'
import path from 'path'

import multer from 'fastify-multer'
import { ulid } from 'ulid'

export const upload = multer({
	storage: multer.diskStorage({
		async destination(req, file, cb) {
			try {
				const folderPath = path.join(tmpdir(), `upload-${ulid()}`)
				await mkdir(folderPath, { recursive: true })

				return cb(null, folderPath)
			} catch (err) {
				return cb(err instanceof Error ? err : new Error(String(err)), '')
			}
		},
		filename(req, file, cb) {
			cb(null, `file${path.extname(file.originalname)}`)
		},
	}),
})
