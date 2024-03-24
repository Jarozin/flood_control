DROP TABLE IF EXISTS flood_record;
CREATE TABLE flood_record (
    id serial primary key,
    user_id int not null,
    date timestamp with time zone not null default now()
);
