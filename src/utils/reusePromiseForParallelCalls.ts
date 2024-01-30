/** Invokes function only once for parallel calls */
export function reusePromiseForParallelCalls<TResult>(
	fn: () => Promise<TResult>,
): () => Promise<TResult> {
	let runningPromise: Promise<TResult> | null = null

	return async function () {
		if (!runningPromise) {
			runningPromise = fn().finally(() => {
				runningPromise = null
			})
		}

		return runningPromise
	}
}
