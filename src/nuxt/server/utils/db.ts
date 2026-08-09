import postgres from "postgres";

const sql = postgres(process.env.DATABASE_URL!, {
	max: 1,
	ssl: process.env.DATABASE_URL?.includes("sslmode=require")
		? "require"
		: false,
});

export { sql };
