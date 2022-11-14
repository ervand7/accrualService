-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
create type "order_status" as enum (
    'NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

create table if not exists "user"
(
    "id"       uuid default gen_random_uuid() primary key,
    "login"    varchar(100)  not null
        unique,
    "password" varchar(100)  not null,
    "token"    varchar(1000) not null
        unique
);

create table if not exists "order"
(
    "id"          bigint                      not null
        unique primary key,
    "user_id"     uuid                        not null,
    "status"      "order_status"              not null,
    "uploaded_at" timestamp without time zone not null default now()
);

alter table only "public"."order"
    add constraint "fk_user_id" foreign key ("user_id")
        references "public"."user" ("id") on delete cascade;

create index "ix_order_status"
    on "public"."order" using btree ("status")
    where ((status = 'NEW'::order_status) OR (status = 'PROCESSING'::order_status));

create table if not exists "accrual"
(
    "order_id" bigint  not null primary key
        unique,
    "user_id"  uuid    not null,
    "amount"   decimal not null
);

alter table only "public"."accrual"
    add constraint "fk_order_id" foreign key ("order_id")
        references "public"."order" ("id") on delete cascade;

create table if not exists "withdrawn"
(
    "order_id"     bigint                      not null primary key
        unique,
    "user_id"      uuid                        not null,
    "amount"       decimal                     not null,
    "processed_at" timestamp without time zone not null default now()
);

alter table only "public"."withdrawn"
    add constraint "fk_order_id" foreign key ("order_id")
        references "public"."order" ("id") on delete cascade;


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
drop table if exists "user" cascade;
drop table if exists "order" cascade;
drop table if exists "accrual" cascade;
drop table if exists "withdrawn" cascade;
drop type if exists "order_status";
