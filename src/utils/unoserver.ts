import timersP from 'node:timers/promises'

import { execa, ExecaChildProcess } from 'execa'
import PQueue from 'p-queue'
import pRetry from 'p-retry'

import { reusePromiseForParallelCalls } from './reusePromiseForParallelCalls.js'

export class Unoserver {
	queue: PQueue
	timeout: number
	unoserver: ExecaChildProcess | null
	port: number

	constructor({
		maxWorkers,
		timeout = 60000,
		port = 12345,
	}: {
		maxWorkers: number
		timeout?: number
		port?: number
	}) {
		this.queue = new PQueue({ concurrency: maxWorkers })
		this.timeout = timeout
		this.unoserver = null
		this.port = port
		this.runServer = reusePromiseForParallelCalls(this.runServer.bind(this))
	}

	private async runServer() {
		const unoserver = execa('unoserver', ['--port', String(this.port)])
		await Promise.race([this.unoserver, await timersP.setTimeout(5000)])
		void unoserver.on('exit', () => {
			this.unoserver = null
		})
		this.unoserver = unoserver
	}

	stopServer(): void {
		if (this.unoserver) {
			this.unoserver.kill()
		}
	}

	/**
	 * Converts source file to target file
	 *
	 * @param from source file
	 * @param to target file
	 */
	convert(from: string, to: string): Promise<void> {
		return this.queue.add(() =>
			pRetry(
				async () => {
					if (!this.unoserver) {
						await this.runServer()
					}

					await execa('unoconvert', ['--port', String(this.port), from, to], {
						timeout: this.timeout,
					})
				},
				{ retries: 3 },
			),
		)
	}
}

export const unoserver = new Unoserver({
	// eslint-disable-next-line @typescript-eslint/strict-boolean-expressions
	maxWorkers: Number(process.env.MAX_WORKERS) || 8,
})
