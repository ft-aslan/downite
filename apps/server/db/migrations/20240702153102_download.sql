-- +goose Up
create table if not exists downloads (
    id integer primary key,
    created_at timestamp default current_timestamp,
    started_at timestamp default null,
    finished_at timestamp default null,
    status int not null,
    name text not null,
    path text not null,
    part_count int not null,
    part_length int not null,
    total_size int not null,
    downloaded_bytes int not null,
    url text not null,
    queue_number int not null,
);
create table if not exists download_parts (
    id integer primary key,
    created_at timestamp default current_timestamp,
    started_at timestamp default null,
    finished_at timestamp default null,
    status int not null,
    download_id int not null,
    part_index int not null,
    start_byte_index int not null,
    end_byte_index int not null,
    downloaded_bytes int not null,
    foreign key (download_id) references downloads (id)
);
-- +goose Down
drop table if exists downloads;
drop table if exists download_parts;
