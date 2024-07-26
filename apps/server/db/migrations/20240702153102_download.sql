-- +goose up
create table if not exists downloads (
    id integer primary key,
    created_at timestamp default current_timestamp,
    started_at timestamp default null,
    finished_at timestamp default null,
    time_active int default 0,
    status int not null,
    name text not null,
    save_path text not null,
    part_count int not null,
    part_length int not null,
    total_size int not null,
    downloaded_bytes int not null,
    url text not null,
    queue_number int not null,
    error text default ''
);

create table if not exists download_parts (
    id integer primary key,
    created_at timestamp default current_timestamp,
    started_at timestamp default null,
    time_active int default 0,
    finished_at timestamp default null,
    status int not null,
    part_index int not null,
    start_byte_index int not null,
    end_byte_index int not null,
    part_length int not null,
    downloaded_bytes int not null,
    error text default '',
    download_id int not null,
    foreign key (download_id) references downloads (id)
);

-- +goose down
drop table downloads;

drop table download_parts;
