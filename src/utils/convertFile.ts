import { extname } from 'path'

import httpErrors from 'http-errors'

import { assert } from './assert.js'
import { unoserver } from './unoserver.js'

/** Converts file to a new format */
export async function convertFile(
	srcPath: string,
	format: string,
): Promise<{ targetPath: string }> {
	const ext = extname(srcPath).toLowerCase()
	assert(
		ext !== '',
		new httpErrors.BadRequest("Can't detect extension for incoming file"),
	)

	let pathWithoutExtension = srcPath.slice(0, srcPath.length - ext.length)

	if (`.${format}` === ext) {
		pathWithoutExtension += '-1'
	}

	const targetPath = `${pathWithoutExtension}.${format}`

	await unoserver.convert(srcPath, targetPath)

	return { targetPath }
}
