import { Ajv } from 'ajv'

// Support multipart/form-data in AJV validation
// https://github.com/fastify/fastify-swagger/issues/47#issuecomment-581986995
export function ajvPlugin(ajv: Ajv): Ajv {
	ajv.addKeyword('isFileType', {
		compile: () => () => true,
	})

	return ajv
}

export const transform = (schema: any): any =>
	schema?.body?.properties
		? {
				...schema,
				body: {
					type: 'object',
					...schema.body,
					properties: schema.body.properties
						? Object.fromEntries(
								Object.entries(schema.body.properties).map(([key, value]) => [
									key,
									(value as any).isFileType ? { type: 'file' } : value,
								]),
						  )
						: undefined,
				},
		  }
		: schema
