DROP TABLE if exists "verify_emails" cascade; 

ALTER TABLE "users" DROP COLUMN "is_email_verified";

