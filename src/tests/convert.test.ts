import { readFileSync } from 'fs'

import { FormData, File } from 'undici'

import { startTestServer } from './startTestServer.js'

const testServer = await startTestServer()

const fixturesFolder = new URL('./fixtures', import.meta.url)
const rtfFile = new File([readFileSync(new URL(`${fixturesFolder.href}/1.rtf`))], '1.rtf')

afterAll(() => testServer.close())

test('/convert/docx', async () => {
	const form = new FormData()
	form.append('file', rtfFile)

	const response = await testServer.request(`convert/docx`, {
		method: 'POST',
		body: form,
	})

	expect(response.headers['content-type']).toBe(
		'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
	)
	expect(response.headers['content-disposition']).toBe('attachment; filename="1.docx"')

	const blob = await response.body.blob()

	const form2 = new FormData()
	form2.append('file', new File([blob], '1.docx'))

	const response2 = await testServer.request(`convert/fodt`, {
		method: 'POST',
		body: form2,
	})
	expect(response2.headers['content-disposition']).toBe('attachment; filename="1.fodt"')

	const text = await response2.body.text()

	expect(text).toMatch(/Hello World!/)
}, 15000)

test('/convert/rtf', async () => {
	const form = new FormData()
	form.append('file', rtfFile)

	const response = await testServer.request(`convert/rtf`, {
		method: 'POST',
		body: form,
	})

	expect(response.headers['content-type']).toBe('application/rtf')
	expect(response.headers['content-disposition']).toBe('attachment; filename="1-1.rtf"')

	const rtf = await response.body.text()

	expect(rtf).toMatch(/Hello World!/)
}, 15000)

test('parallel convertion', async () => {
	const results = await Promise.allSettled(
		Array(100)
			.fill(null)
			.map(async () => {
				const form = new FormData()
				form.append('file', rtfFile)
				return testServer.request(`convert/pdf`, { method: 'POST', body: form })
			}),
	)

	const error = results.find(
		(result): result is PromiseRejectedResult => result.status === 'rejected',
	)

	if (error) {
		throw error.reason
	}
}, 30000)
