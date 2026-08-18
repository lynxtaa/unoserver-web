import { randomUUID } from 'node:crypto'
import { mkdir } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'

export async function createRandomFolder(): Promise<string> {
	const folderPath = path.join(tmpdir(), `upload-${randomUUID()}`)
	await mkdir(folderPath, { recursive: true })
	return folderPath
}
