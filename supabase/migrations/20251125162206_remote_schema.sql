


SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


COMMENT ON SCHEMA "public" IS 'standard public schema';



CREATE EXTENSION IF NOT EXISTS "pg_graphql" WITH SCHEMA "graphql";






CREATE EXTENSION IF NOT EXISTS "pg_stat_statements" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "extensions";






CREATE EXTENSION IF NOT EXISTS "supabase_vault" WITH SCHEMA "vault";






CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA "extensions";






CREATE TYPE "public"."event_type" AS ENUM (
    'error',
    'perf',
    'user_action'
);


ALTER TYPE "public"."event_type" OWNER TO "postgres";


CREATE TYPE "public"."member_role" AS ENUM (
    'viewer',
    'admin',
    'owner'
);


ALTER TYPE "public"."member_role" OWNER TO "postgres";


CREATE TYPE "public"."severity_level" AS ENUM (
    'warn',
    'error',
    'fatal'
);


ALTER TYPE "public"."severity_level" OWNER TO "postgres";

SET default_tablespace = '';

SET default_table_access_method = "heap";


CREATE TABLE IF NOT EXISTS "public"."audit_logs" (
    "id" bigint NOT NULL,
    "project_id" "uuid" NOT NULL,
    "actor_type" "text" NOT NULL,
    "actor_id" "text",
    "action" "text" NOT NULL,
    "success" boolean NOT NULL,
    "status_code" integer NOT NULL,
    "auth_mode" "text" NOT NULL,
    "details" "jsonb",
    "request_id" "text",
    "ip_hash" "text",
    "user_agent" "text",
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "audit_logs_actor_type_check" CHECK (("actor_type" = ANY (ARRAY['admin'::"text", 'user'::"text", 'system'::"text"]))),
    CONSTRAINT "audit_logs_auth_mode_check" CHECK (("auth_mode" = ANY (ARRAY['PAT'::"text", 'HMAC'::"text", 'PK_ONLY'::"text", 'SYSTEM'::"text"])))
);


ALTER TABLE "public"."audit_logs" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."audit_logs_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."audit_logs_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."audit_logs_id_seq" OWNED BY "public"."audit_logs"."id";



CREATE TABLE IF NOT EXISTS "public"."events" (
    "project_id" "uuid" NOT NULL,
    "event_id" "text" NOT NULL,
    "event_type" "public"."event_type" NOT NULL,
    "event_ts" timestamp with time zone NOT NULL,
    "ingested_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "op" "text" NOT NULL,
    "variant" "text",
    "surface" "text",
    "version" "text" NOT NULL,
    "args_sig" "text",
    "args_count" integer,
    "success" boolean,
    "severity" "public"."severity_level",
    "code" "text",
    "message" "text",
    "stack" "text",
    "duration_ms" integer,
    "user_id_h" "text" NOT NULL,
    "session_id" "text",
    "trace_id" "text",
    "payload" "jsonb",
    "sdk_name" "text" NOT NULL,
    "sdk_version" "text" NOT NULL,
    "sdk_language" "text",
    "sdk_runtime" "text",
    "sdk_payload" "jsonb"
);


ALTER TABLE "public"."events" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."project_keys" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "project_id" "uuid" NOT NULL,
    "label" "text" NOT NULL,
    "env" "text" NOT NULL,
    "signed_only" boolean DEFAULT false NOT NULL,
    "public_key" "text" NOT NULL,
    "secret_enc" "text",
    "secret_fingerprint" "text",
    "disabled" boolean DEFAULT false NOT NULL,
    "created_by" "uuid" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "last_used_at" timestamp with time zone,
    CONSTRAINT "project_keys_env_check" CHECK (("env" = ANY (ARRAY['prod'::"text", 'staging'::"text", 'dev'::"text"]))),
    CONSTRAINT "project_keys_label_check" CHECK ((("char_length"("label") >= 1) AND ("char_length"("label") <= 64)))
);


ALTER TABLE "public"."project_keys" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."project_members" (
    "project_id" "uuid" NOT NULL,
    "user_id" "uuid" NOT NULL,
    "role" "public"."member_role" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL
);


ALTER TABLE "public"."project_members" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."project_tokens" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "project_id" "uuid" NOT NULL,
    "label" "text" NOT NULL,
    "scopes" "text"[] NOT NULL,
    "token_hash" "text" NOT NULL,
    "created_by" "uuid" NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "last_used_at" timestamp with time zone,
    "expires_at" timestamp with time zone,
    "revoked" boolean DEFAULT false NOT NULL,
    CONSTRAINT "project_tokens_label_check" CHECK ((("char_length"("label") >= 1) AND ("char_length"("label") <= 64)))
);


ALTER TABLE "public"."project_tokens" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."projects" (
    "id" "uuid" DEFAULT "gen_random_uuid"() NOT NULL,
    "name" "text" NOT NULL,
    "owner_user_id" "uuid" NOT NULL,
    "retention_days" integer,
    "signed_only" boolean DEFAULT false NOT NULL,
    "created_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    "is_demo" boolean DEFAULT false NOT NULL,
    CONSTRAINT "projects_name_check" CHECK ((("char_length"("name") >= 1) AND ("char_length"("name") <= 80)))
);


ALTER TABLE "public"."projects" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."user_consent" (
    "project_id" "uuid" NOT NULL,
    "user_id_h" "text" NOT NULL,
    "state" "text" NOT NULL,
    "updated_at" timestamp with time zone DEFAULT "now"() NOT NULL,
    CONSTRAINT "user_consent_state_check" CHECK (("state" = ANY (ARRAY['granted'::"text", 'revoked'::"text"]))),
    CONSTRAINT "user_consent_user_id_check" CHECK (("char_length"("user_id_h") <= 128))
);


ALTER TABLE "public"."user_consent" OWNER TO "postgres";


CREATE TABLE IF NOT EXISTS "public"."user_consent_history" (
    "id" bigint NOT NULL,
    "project_id" "uuid" NOT NULL,
    "user_id_h" "text" NOT NULL,
    "prev_state" "text",
    "next_state" "text" NOT NULL,
    "ts" timestamp with time zone DEFAULT "now"() NOT NULL,
    "version" "text",
    "meta" "jsonb",
    CONSTRAINT "user_consent_history_next_state_check" CHECK (("next_state" = ANY (ARRAY['granted'::"text", 'revoked'::"text"])))
);


ALTER TABLE "public"."user_consent_history" OWNER TO "postgres";


CREATE SEQUENCE IF NOT EXISTS "public"."user_consent_history_id_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE "public"."user_consent_history_id_seq" OWNER TO "postgres";


ALTER SEQUENCE "public"."user_consent_history_id_seq" OWNED BY "public"."user_consent_history"."id";



ALTER TABLE ONLY "public"."audit_logs" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."audit_logs_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."user_consent_history" ALTER COLUMN "id" SET DEFAULT "nextval"('"public"."user_consent_history_id_seq"'::"regclass");



ALTER TABLE ONLY "public"."audit_logs"
    ADD CONSTRAINT "audit_logs_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."events"
    ADD CONSTRAINT "events_pk" PRIMARY KEY ("project_id", "event_id");



ALTER TABLE ONLY "public"."project_keys"
    ADD CONSTRAINT "project_keys_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."project_members"
    ADD CONSTRAINT "project_members_pkey" PRIMARY KEY ("project_id", "user_id");



ALTER TABLE ONLY "public"."project_tokens"
    ADD CONSTRAINT "project_tokens_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."projects"
    ADD CONSTRAINT "projects_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."user_consent_history"
    ADD CONSTRAINT "user_consent_history_pkey" PRIMARY KEY ("id");



ALTER TABLE ONLY "public"."user_consent"
    ADD CONSTRAINT "user_consent_pkey" PRIMARY KEY ("project_id", "user_id_h");



CREATE INDEX "idx_audit_logs_action" ON "public"."audit_logs" USING "btree" ("action");



CREATE INDEX "idx_audit_logs_project_id_created_at" ON "public"."audit_logs" USING "btree" ("project_id", "created_at" DESC);



CREATE INDEX "idx_events_project_id_event_ts" ON "public"."events" USING "btree" ("project_id", "event_ts" DESC);



CREATE INDEX "idx_events_project_id_ingested_at" ON "public"."events" USING "btree" ("project_id", "ingested_at" DESC);



CREATE INDEX "idx_events_session_id" ON "public"."events" USING "btree" ("session_id");



CREATE INDEX "idx_events_trace_id" ON "public"."events" USING "btree" ("trace_id");



CREATE INDEX "idx_events_user_id_h" ON "public"."events" USING "btree" ("user_id_h");



CREATE INDEX "idx_project_keys_project_id_disabled" ON "public"."project_keys" USING "btree" ("project_id", "disabled");



CREATE UNIQUE INDEX "idx_project_keys_public_key" ON "public"."project_keys" USING "btree" ("public_key");



CREATE INDEX "idx_project_members_user_id_role" ON "public"."project_members" USING "btree" ("user_id", "role");



CREATE INDEX "idx_project_tokens_project_id_revoked" ON "public"."project_tokens" USING "btree" ("project_id", "revoked");



CREATE INDEX "idx_project_tokens_scopes_gin" ON "public"."project_tokens" USING "gin" ("scopes");



CREATE INDEX "idx_projects_owner_user_id" ON "public"."projects" USING "btree" ("owner_user_id");



CREATE INDEX "idx_user_consent_history_project_id_ts_desc" ON "public"."user_consent_history" USING "btree" ("project_id", "ts" DESC);



ALTER TABLE ONLY "public"."audit_logs"
    ADD CONSTRAINT "audit_logs_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."events"
    ADD CONSTRAINT "events_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."project_keys"
    ADD CONSTRAINT "project_keys_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "auth"."users"("id");



ALTER TABLE ONLY "public"."project_keys"
    ADD CONSTRAINT "project_keys_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."project_members"
    ADD CONSTRAINT "project_members_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."project_members"
    ADD CONSTRAINT "project_members_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "auth"."users"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."project_tokens"
    ADD CONSTRAINT "project_tokens_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "auth"."users"("id");



ALTER TABLE ONLY "public"."project_tokens"
    ADD CONSTRAINT "project_tokens_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."projects"
    ADD CONSTRAINT "projects_owner_user_id_fkey" FOREIGN KEY ("owner_user_id") REFERENCES "auth"."users"("id") ON DELETE RESTRICT;



ALTER TABLE ONLY "public"."user_consent_history"
    ADD CONSTRAINT "user_consent_history_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON DELETE CASCADE;



ALTER TABLE ONLY "public"."user_consent"
    ADD CONSTRAINT "user_consent_project_id_fkey" FOREIGN KEY ("project_id") REFERENCES "public"."projects"("id") ON DELETE CASCADE;



CREATE POLICY "Admins and owners can add members" ON "public"."project_members" FOR INSERT WITH CHECK ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_members"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can create keys" ON "public"."project_keys" FOR INSERT WITH CHECK ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_keys"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can create tokens" ON "public"."project_tokens" FOR INSERT WITH CHECK ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_tokens"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can delete keys" ON "public"."project_keys" FOR DELETE USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_keys"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can remove members" ON "public"."project_members" FOR DELETE USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_members"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can update keys" ON "public"."project_keys" FOR UPDATE USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_keys"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can update members" ON "public"."project_members" FOR UPDATE USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_members"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can update projects" ON "public"."projects" FOR UPDATE USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "projects"."id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can update tokens" ON "public"."project_tokens" FOR UPDATE USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_tokens"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can view audit logs" ON "public"."audit_logs" FOR SELECT USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "audit_logs"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Admins and owners can view keys" ON "public"."project_keys" FOR SELECT USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_keys"."project_id") AND ("pm"."user_id" = "auth"."uid"()) AND ("pm"."role" = ANY (ARRAY['admin'::"public"."member_role", 'owner'::"public"."member_role"]))))));



CREATE POLICY "Authenticated users can create projects" ON "public"."projects" FOR INSERT WITH CHECK (("auth"."uid"() = "owner_user_id"));



CREATE POLICY "Members can view events" ON "public"."events" FOR SELECT USING (((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "events"."project_id") AND ("pm"."user_id" = "auth"."uid"())))) OR (EXISTS ( SELECT 1
   FROM "public"."projects" "p"
  WHERE (("p"."id" = "events"."project_id") AND ("p"."is_demo" = true))))));



CREATE POLICY "Members can view project members" ON "public"."project_members" FOR SELECT USING ((EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "project_members"."project_id") AND ("pm"."user_id" = "auth"."uid"())))));



CREATE POLICY "Members can view projects" ON "public"."projects" FOR SELECT USING ((("is_demo" = true) OR (EXISTS ( SELECT 1
   FROM "public"."project_members" "pm"
  WHERE (("pm"."project_id" = "projects"."id") AND ("pm"."user_id" = "auth"."uid"()))))));



CREATE POLICY "Only owners can delete projects" ON "public"."projects" FOR DELETE USING (("owner_user_id" = "auth"."uid"()));



CREATE POLICY "Only service role can delete consent" ON "public"."user_consent" FOR DELETE USING (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can delete consent history" ON "public"."user_consent_history" FOR DELETE USING (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can insert consent" ON "public"."user_consent" FOR INSERT WITH CHECK (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can insert consent history" ON "public"."user_consent_history" FOR INSERT WITH CHECK (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can insert events" ON "public"."events" FOR INSERT WITH CHECK (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can select consent" ON "public"."user_consent" FOR SELECT USING (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can select consent history" ON "public"."user_consent_history" FOR SELECT USING (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can update consent" ON "public"."user_consent" FOR UPDATE USING (("auth"."role"() = 'service_role'::"text")) WITH CHECK (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Only service role can update consent history" ON "public"."user_consent_history" FOR UPDATE USING (("auth"."role"() = 'service_role'::"text")) WITH CHECK (("auth"."role"() = 'service_role'::"text"));



CREATE POLICY "Service role can insert audit logs" ON "public"."audit_logs" FOR INSERT WITH CHECK (("auth"."role"() = 'service_role'::"text"));



ALTER TABLE "public"."audit_logs" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."events" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."project_keys" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."project_members" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."project_tokens" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."projects" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."user_consent" ENABLE ROW LEVEL SECURITY;


ALTER TABLE "public"."user_consent_history" ENABLE ROW LEVEL SECURITY;




ALTER PUBLICATION "supabase_realtime" OWNER TO "postgres";


GRANT USAGE ON SCHEMA "public" TO "postgres";
GRANT USAGE ON SCHEMA "public" TO "anon";
GRANT USAGE ON SCHEMA "public" TO "authenticated";
GRANT USAGE ON SCHEMA "public" TO "service_role";








































































































































































GRANT ALL ON TABLE "public"."audit_logs" TO "anon";
GRANT ALL ON TABLE "public"."audit_logs" TO "authenticated";
GRANT ALL ON TABLE "public"."audit_logs" TO "service_role";



GRANT ALL ON SEQUENCE "public"."audit_logs_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."audit_logs_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."audit_logs_id_seq" TO "service_role";



GRANT ALL ON TABLE "public"."events" TO "anon";
GRANT ALL ON TABLE "public"."events" TO "authenticated";
GRANT ALL ON TABLE "public"."events" TO "service_role";



GRANT ALL ON TABLE "public"."project_keys" TO "anon";
GRANT ALL ON TABLE "public"."project_keys" TO "authenticated";
GRANT ALL ON TABLE "public"."project_keys" TO "service_role";



GRANT ALL ON TABLE "public"."project_members" TO "anon";
GRANT ALL ON TABLE "public"."project_members" TO "authenticated";
GRANT ALL ON TABLE "public"."project_members" TO "service_role";



GRANT ALL ON TABLE "public"."project_tokens" TO "anon";
GRANT ALL ON TABLE "public"."project_tokens" TO "authenticated";
GRANT ALL ON TABLE "public"."project_tokens" TO "service_role";



GRANT ALL ON TABLE "public"."projects" TO "anon";
GRANT ALL ON TABLE "public"."projects" TO "authenticated";
GRANT ALL ON TABLE "public"."projects" TO "service_role";



GRANT ALL ON TABLE "public"."user_consent" TO "anon";
GRANT ALL ON TABLE "public"."user_consent" TO "authenticated";
GRANT ALL ON TABLE "public"."user_consent" TO "service_role";



GRANT ALL ON TABLE "public"."user_consent_history" TO "anon";
GRANT ALL ON TABLE "public"."user_consent_history" TO "authenticated";
GRANT ALL ON TABLE "public"."user_consent_history" TO "service_role";



GRANT ALL ON SEQUENCE "public"."user_consent_history_id_seq" TO "anon";
GRANT ALL ON SEQUENCE "public"."user_consent_history_id_seq" TO "authenticated";
GRANT ALL ON SEQUENCE "public"."user_consent_history_id_seq" TO "service_role";









ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON SEQUENCES TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON FUNCTIONS TO "service_role";






ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "postgres";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "anon";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "authenticated";
ALTER DEFAULT PRIVILEGES FOR ROLE "postgres" IN SCHEMA "public" GRANT ALL ON TABLES TO "service_role";































