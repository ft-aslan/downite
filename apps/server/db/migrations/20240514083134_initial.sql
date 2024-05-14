-- +goose up
create table
    if not exists users (
        id int primary key,
        created_at timestamp default current_timestamp,
        username text unique not null,
        password text not null
    );

create table
    if not exists torrents (
        id int primary key,
        infohash text not null unique,
        created_at timestamp default current_timestamp,
        name text not null,
        save_path text not null,
        status int not null,
        time_active timestamp,
        downloaded int,
        uploaded int,
        total_size int,
        comment text,
        category_id int,
        foreign key (category_id) references categories (id)
    );

create table
    if not exists files (
        id int primary key,
        torrent_id int not null,
        created_at timestamp default current_timestamp,
        name text not null,
        length int not null,
        path text not null,
        priority int not null,
        piece_index int not null,
        foreign key (torrent_id) references torrents (id)
    );

create table
    if not exists trackers (
        id int primary key,
        created_at timestamp default current_timestamp,
        address text not null unique
    );

create table
    if not exists torrent_trackers (
        id int primary key,
        created_at timestamp default current_timestamp,
        torrent_id int not null,
        tracker_id int not null,
        foreign key (torrent_id) references torrents (id),
        foreign key (tracker_id) references trackers (id)
    );

create table
    if not exists tags (
        id int primary key,
        created_at timestamp default current_timestamp,
        name text not null
    );

create table
    if not exists torrent_tags (
        id int primary key,
        created_at timestamp default current_timestamp,
        torrent_id int not null,
        tag_id int not null,
        foreign key (torrent_id) references torrents (id),
        foreign key (tag_id) references tags (id)
    );

create table
    if not exists categories (
        id serial primary key,
        created_at timestamp default current_timestamp,
        name text not null,
        save_path text not null,
        incomplete_save_path text
    );

-- +goose down
drop table users;

drop table torrents;

drop table files;

drop table trackers;

drop table torrent_trackers;

drop table tags;

drop table torrent_tags;

drop table categories;