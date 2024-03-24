DROP TABLE IF EXISTS flood_record;
CREATE TABLE flood_record (
    id serial primary key,
    user_id int not null,
    date timestamp not null default now()
);
