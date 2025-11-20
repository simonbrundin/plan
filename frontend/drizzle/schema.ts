import { pgTable, unique, integer, varchar, timestamp, foreignKey, primaryKey } from "drizzle-orm/pg-core"
import { sql } from "drizzle-orm"



export const users = pgTable("users", {
	id: integer().primaryKey().generatedByDefaultAsIdentity({ name: "users_id_seq", startWith: 1, increment: 1, minValue: 1, maxValue: 2147483647, cache: 1 }),
	sub: varchar({ length: 255 }).notNull(),
	email: varchar({ length: 40 }).notNull(),
	firstName: varchar("first_name", { length: 16 }),
	lastName: varchar("last_name", { length: 40 }),
	created: timestamp({ withTimezone: true, mode: 'string' }).default(sql`CURRENT_TIMESTAMP`).notNull(),
}, (table) => [
	unique("users_sub_key").on(table.sub),
	unique("users_email_key").on(table.email),
]);

export const goals = pgTable("goals", {
	id: integer().primaryKey().generatedByDefaultAsIdentity({ name: "goals_id_seq", startWith: 1, increment: 1, minValue: 1, maxValue: 2147483647, cache: 1 }),
	title: varchar({ length: 64 }),
	created: timestamp({ withTimezone: true, mode: 'string' }).default(sql`CURRENT_TIMESTAMP`).notNull(),
	finished: timestamp({ withTimezone: true, mode: 'string' }),
	icon: varchar({ length: 100 }).default('roentgen:default'),
});

export const userGoals = pgTable("user_goals", {
	userId: integer("user_id").notNull(),
	goalId: integer("goal_id").notNull(),
}, (table) => [
	foreignKey({
			columns: [table.userId],
			foreignColumns: [users.id],
			name: "user_goals_user_id_fkey"
		}),
	foreignKey({
			columns: [table.goalId],
			foreignColumns: [goals.id],
			name: "user_goals_goal_id_fkey"
		}),
	primaryKey({ columns: [table.userId, table.goalId], name: "user_goals_pkey"}),
]);

export const goalRelations = pgTable("goal_relations", {
	parentId: integer("parent_id").notNull(),
	childId: integer("child_id").notNull(),
	order: integer().default(0).notNull(),
	weight: integer().default(10).notNull(),
}, (table) => [
	foreignKey({
			columns: [table.parentId],
			foreignColumns: [goals.id],
			name: "goal_relations_parent_id_fkey"
		}),
	foreignKey({
			columns: [table.childId],
			foreignColumns: [goals.id],
			name: "goal_relations_child_id_fkey"
		}),
	primaryKey({ columns: [table.parentId, table.childId], name: "goal_relations_pkey"}),
]);
