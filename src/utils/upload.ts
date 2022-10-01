import { randomUUID } from 'node:crypto'
import { mkdir } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'

import multer from 'fastify-multer'

async function createRandomFolder() {
	const folderPath = path.join(tmpdir(), `upload-${randomUUID()}`)
	await mkdir(folderPath, { recursive: true })
	return folderPath
}

export const upload = multer({
	storage: multer.diskStorage({
		destination(req, file, cb) {
			createRandomFolder()
				.then(folderPath => {
					cb(null, folderPath)
				})
				.catch(err => {
					cb(err instanceof Error ? err : new Error(String(err)), '')
				})
		},
		filename(req, file, cb) {
			cb(null, file.originalname)
		},
	}),
})
