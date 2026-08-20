import { openAsBlob } from 'node:fs'

import { FormData } from 'undici'

import { startTestServer } from './startTestServer.js'

const testServer = await startTestServer()

const rtfBlob = await openAsBlob('./src/tests/fixtures/1.rtf')

const rtfFile = new File([rtfBlob], '1.rtf')

afterAll(async () => testServer.close())

test('/convert/docx', async () => {
	const form = new FormData()
	form.append('file', rtfFile)

	const response = await testServer.fetch(`convert/docx`, {
		method: 'POST',
		body: form,
	})

	expect(response.headers.get('content-type')).toBe(
		'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
	)
	expect(response.headers.get('content-disposition')).toBe('attachment; filename=1.docx')

	const blob = await response.blob()

	{
		const form = new FormData()
		form.append('file', new File([blob], '1.docx'))

		const response = await testServer.fetch(`convert/fodt`, {
			method: 'POST',
			body: form,
		})
		expect(response.headers.get('content-disposition')).toBe(
			'attachment; filename=1.fodt',
		)

		const text = await response.text()

		expect(text).toMatch(/Hello World!/)
	}
}, 30000)

test('/convert/rtf', async () => {
	const form = new FormData()
	form.append('file', rtfFile)

	const response = await testServer.fetch(`convert/rtf`, {
		method: 'POST',
		body: form,
	})

	expect(response.headers.get('content-type')).toBe('application/rtf')
	expect(response.headers.get('content-disposition')).toBe('attachment; filename=1.rtf')

	const rtf = await response.text()

	expect(rtf).toMatch(/Hello World!/)
}, 15000)

test('converts files bigger than 1 MiB', async () => {
	const paragraph = '\\pard\\sa200\\sl276\\slmult1\\f0\\fs22\\lang9 Hello World!\\par\n'
	const bigRtf = `{\\rtf1\\ansi\\deff0{\\fonttbl{\\f0\\fnil\\fcharset0 Calibri;}}\n${paragraph.repeat(30_000)}}`

	expect(bigRtf.length).toBeGreaterThan(1024 * 1024)

	const form = new FormData()
	form.append('file', new File([bigRtf], 'big.rtf'))

	const response = await testServer.fetch(`convert/rtf`, {
		method: 'POST',
		body: form,
	})

	expect(response.status).toBe(200)
	await response.blob()
}, 60000)

test('parallel convertion', async () => {
	const results = await Promise.allSettled(
		Array.from({ length: 64 }, () => null).map(async () => {
			const form = new FormData()
			form.append('file', rtfFile)
			return testServer.fetch(`convert/pdf`, { method: 'POST', body: form })
		}),
	)

	const error = results.find(
		(result): result is PromiseRejectedResult => result.status === 'rejected',
	)

	if (error) {
		throw error.reason
	}
}, 30000)
