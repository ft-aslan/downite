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
        comment text
    );

create table
    if not exists files (
        id serial primary key,
        torrent_id int,
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
        id serial primary key,
        created_at timestamp default current_timestamp,
        address text not null unique
    );

create table
    if not exists torrent_trackers (
        id serial primary key,
        created_at timestamp default current_timestamp,
        torrent_id int,
        tracker_id int,
        foreign key (torrent_id) references torrents (id),
        foreign key (tracker_id) references trackers (id)
    );

create table
    if not exists tags (
        id serial primary key,
        created_at timestamp default current_timestamp,
        name text not null
    );

create table
    if not exists torrent_tags (
        id serial primary key,
        created_at timestamp default current_timestamp,
        torrent_id int,
        tag_id int,
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

create table
    if not exists torrent_categories (
        id serial primary key,
        created_at timestamp default current_timestamp,
        torrent_id int,
        category_id int,
        foreign key (torrent_id) references torrents (id),
        foreign key (category_id) references categories (id)
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

drop table torrent_categories;