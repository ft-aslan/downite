select
    torrents.id,
    torrents.infohash,
    torrents.name,
    torrents.save_path,
    torrents.status,
    torrents.time_active,
    torrents.downloaded,
    torrents.uploaded,
    torrents.total_size,
    torrents.comment,
    torrents.category_id,
    torrents.created_at,
    torrents.started_at,
    group_concat (tags.name) as tags,
    group_concat (trackers.address) as trackers
from
    torrents
    left join torrent_tags on torrent_tags.torrent_id = torrents.id
    left join tags on tags.id = torrent_tags.tag_id
    left join torrent_trackers on torrent_trackers.torrent_id = torrents.id
    left join trackers on trackers.rowid = torrent_trackers.tracker_id
group by
    torrents.id;

select
    torrents.id,
    torrents.infohash,
    torrents.name,
    torrents.save_path,
    torrents.status,
    torrents.time_active,
    torrents.downloaded,
    torrents.uploaded,
    torrents.total_size,
    torrents.comment,
    torrents.category_id,
    torrents.created_at
from
    torrents
order by
    created_at;

/* SELECT
tags.name
FROM
tags
JOIN torrent_tags ON torrent_tags.tag_id = tags.id
WHERE
torrent_tags.torrent_id = 1 */