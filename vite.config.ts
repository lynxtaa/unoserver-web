import { defineConfig } from 'vitest/config'

const isCi = process.env.CI !== undefined

export default defineConfig({
	test: {
		mockReset: true,
		environment: 'node',
		root: 'src',
		include: ['**/*.test.{ts,tsx}'],
		globals: true,
		coverage: {
			provider: 'c8',
			reporter: ['html', isCi ? 'text' : 'text-summary'],
			all: true,
			include: ['**/*.ts'],
		},
		snapshotFormat: {
			escapeString: false,
			printBasicPrototype: false,
		},
	},
})
