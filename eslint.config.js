import eslintConfig from '@lynxtaa/eslint-config'
import esm from '@lynxtaa/eslint-config/esm'
import requiresTypechecking from '@lynxtaa/eslint-config/requires-typechecking'

export default [
	...eslintConfig,
	...esm,
	...requiresTypechecking,
	// See https://typescript-eslint.io/getting-started/typed-linting
	{
		languageOptions: {
			parserOptions: {
				projectService: {
					allowDefaultProject: ['eslint.config.js', 'vite.config.ts'],
				},
				tsconfigRootDir: import.meta.dirname,
			},
		},
	},
]
