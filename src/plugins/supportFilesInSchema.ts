export const transform = (data: any): any =>
	data.schema?.body?.properties
		? {
				...data,
				schema: {
					...data.schema,
					body: {
						type: 'object',
						...data.schema.body,
						properties: data.schema.body.properties
							? Object.fromEntries(
									Object.entries(data.schema.body.properties).map(([key, value]) => [
										key,
										(value as any).format === 'binary' ? { type: 'file' } : value,
									]),
							  )
							: undefined,
					},
				},
		  }
		: data
