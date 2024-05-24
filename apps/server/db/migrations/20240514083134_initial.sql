-- +goose up
create table
    if not exists users (
        id integer primary key,
        created_at timestamp default current_timestamp,
        username text unique not null,
        password text not null
    );

create table
    if not exists torrents (
        infohash text primary key,
        created_at timestamp default current_timestamp,
        started_at timestamp,
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
        id integer primary key,
        infohash int not null,
        created_at timestamp default current_timestamp,
        name text not null,
        path text not null,
        priority int not null,
        foreign key (infohash) references torrents (infohash)
    );

create table
    if not exists trackers (
        id integer primary key,
        created_at timestamp default current_timestamp,
        url text not null unique
    );

create table
    if not exists torrent_trackers (
        id integer primary key,
        created_at timestamp default current_timestamp,
        infohash int not null,
        tracker_id int not null,
        tier int not null,
        foreign key (infohash) references torrents (infohash),
        foreign key (tracker_id) references trackers (id)
    );

create table
    if not exists tags (
        id integer primary key,
        created_at timestamp default current_timestamp,
        name text not null
    );

create table
    if not exists torrent_tags (
        id integer primary key,
        created_at timestamp default current_timestamp,
        infohash int not null,
        tag_id int not null,
        foreign key (infohash) references torrents (infohash),
        foreign key (tag_id) references tags (id)
    );

create table
    if not exists categories (
        id integer primary key,
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