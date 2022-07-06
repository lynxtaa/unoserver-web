import { mkdir } from 'fs/promises'
import { tmpdir } from 'os'
import path from 'path'

import multer from 'fastify-multer'
import { ulid } from 'ulid'

async function createRandomFolder() {
	const folderPath = path.join(tmpdir(), `upload-${ulid()}`)
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
