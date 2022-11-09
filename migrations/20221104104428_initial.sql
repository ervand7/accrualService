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
    "id"          uuid                                 default gen_random_uuid() primary key,
    "user_id"     uuid                        not null,
    "number"      bigint                      not null
        unique,
    "status"      "order_status"              not null,
    "uploaded_at" timestamp without time zone not null default now()
);

create table if not exists "accrual"
(
    "id"       uuid default gen_random_uuid() primary key,
    "user_id"  uuid    not null,
    "order_id" uuid    not null,
    "amount"   decimal not null
);

create table if not exists "withdrawal"
(
    "id"           uuid                                 default gen_random_uuid() primary key,
    "user_id"      uuid                        not null,
    "order_id"     uuid                        not null,
    "amount"       decimal                     not null,
    "processed_at" timestamp without time zone not null default now()
);


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
drop table if exists "auth";
drop table if exists "user";
drop table if exists "order";
drop table if exists "accrual";
drop table if exists "withdrawal";
drop type if exists "order_status";
