import postgres from "postgres";

const sql = postgres(process.env.DATABASE_URL!, {
	max: 10,
	transform: {
		undefined: null,
	},
});

export { sql };
