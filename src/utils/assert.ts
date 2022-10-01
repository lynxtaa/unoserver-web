export function assert(cond: boolean, msg: string | Error): asserts cond {
	if (!cond) {
		throw typeof msg === 'string' ? new Error(msg) : msg
	}
}
