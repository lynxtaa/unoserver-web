import timersP from 'node:timers/promises'

import { execa, type ExecaChildProcess } from 'execa'
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
		await Promise.race([this.unoserver, timersP.setTimeout(5000)])
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
	 * @param options conversion options
	 */
	async convert(from: string, to: string, options?: { filter?: string }): Promise<void> {
		return this.queue.add(async () =>
			pRetry(
				async () => {
					if (!this.unoserver) {
						await this.runServer()
					}

					const portCommandArg = ['--port', String(this.port)]
					const filterCommandArg =
						options?.filter !== undefined ? ['--filter', options.filter] : []

					const commandArguments = [...portCommandArg, ...filterCommandArg, from, to]

					await execa('unoconvert', commandArguments, {
						timeout: this.timeout,
					})
				},
				{
					retries:
						process.env.CONVERSION_RETRIES !== undefined
							? Number(process.env.CONVERSION_RETRIES)
							: 3,
				},
			),
		)
	}
}

export const unoserver = new Unoserver({
	maxWorkers: process.env.MAX_WORKERS !== undefined ? Number(process.env.MAX_WORKERS) : 8,
})
