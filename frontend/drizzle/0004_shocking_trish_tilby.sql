DROP TABLE "credentials" CASCADE;--> statement-breakpoint
ALTER TABLE "goal_relations" ADD COLUMN "order" integer DEFAULT 0 NOT NULL;--> statement-breakpoint
ALTER TABLE "goal_relations" ADD COLUMN "weight" integer DEFAULT 10 NOT NULL;--> statement-breakpoint
ALTER TABLE "goals" ADD COLUMN "icon" varchar(100) DEFAULT 'roentgen:default';