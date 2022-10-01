module.exports = {
	extends: [
		'@lynxtaa/eslint-config',
		'@lynxtaa/eslint-config/esm',
		'@lynxtaa/eslint-config/requires-typechecking',
	],
	parserOptions: {
		tsconfigRootDir: __dirname,
		project: ['./tsconfig.json'],
	},
}
