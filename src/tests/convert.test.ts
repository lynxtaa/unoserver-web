import { readFileSync } from 'fs'

import { FormData, File } from 'formdata-node'

import { startTestServer } from './startTestServer.js'

const testServer = await startTestServer()

const fixturesFolder = new URL('./fixtures', import.meta.url)
const rtfFile = new File([readFileSync(new URL(`${fixturesFolder}/1.rtf`))], '1.rtf')

afterAll(() => testServer.close())

test('/convert/docx', async () => {
	const form = new FormData()
	form.append('file', rtfFile)

	const response = await testServer.fetch(`convert/docx`, {
		method: 'POST',
		body: form as any,
	})

	expect(response.headers.get('content-type')).toBe(
		'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
	)
	expect(response.headers.get('content-disposition')).toBe(
		'attachment; filename="1.docx"',
	)

	const buffer = await response.arrayBuffer()

	const form2 = new FormData()
	form2.append('file', new File([buffer], '1.docx'))

	const response2 = await testServer.fetch(`convert/fodt`, {
		method: 'POST',
		body: form2 as any,
	})
	expect(response2.headers.get('content-disposition')).toBe(
		'attachment; filename="1.fodt"',
	)

	const text = await response2.text()

	expect(text).toMatch(/Hello World!/)
}, 15000)

test('/convert/rtf', async () => {
	const form = new FormData()
	form.append('file', rtfFile)

	const response = await testServer.fetch(`convert/rtf`, {
		method: 'POST',
		body: form as any,
	})

	expect(response.headers.get('content-type')).toBe('application/rtf')
	expect(response.headers.get('content-disposition')).toBe(
		'attachment; filename="1-1.rtf"',
	)

	const rtf = await response.text()

	expect(rtf).toMatch(/Hello World!/)
}, 15000)

test('parallel convertion', async () => {
	const results = await Promise.allSettled(
		Array(100)
			.fill(null)
			.map(async () => {
				const form = new FormData()
				form.append('file', rtfFile)
				return testServer.fetch(`convert/pdf`, { method: 'POST', body: form as any })
			}),
	)

	const error = results.find(
		(result): result is PromiseRejectedResult => result.status === 'rejected',
	)

	if (error) {
		throw error.reason
	}
}, 15000)
