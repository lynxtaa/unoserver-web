import { defineConfig } from 'vitest/config'

const isCi = process.env.CI !== undefined

export default defineConfig({
	test: {
		mockReset: true,
		environment: 'node',
		include: ['src/**/*.test.{ts,tsx}'],
		globals: true,
		coverage: {
			provider: 'v8',
			reporter: ['html', isCi ? 'text' : 'text-summary'],
			all: true,
			include: ['src/**/*.ts'],
			reportsDirectory: 'coverage',
		},
		snapshotFormat: {
			escapeString: false,
			printBasicPrototype: false,
		},
	},
})
