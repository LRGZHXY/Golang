create table `videos`(
                       id bigint primary key auto_increment,
                       video_id varchar(20) not null
);

insert into `videos` (video_id)
values
    ('7482354135765880091'),
    ('7477984552787397946'),
    ('7474583086403980604'),
    ('7461981146180537641'),
    ('7481509204755369270'),
    ('7480744910195363124'),
    ('7456408637376974140'),
    ('7475985216935415094'),
    ('7474558070971010355'),
    ('7478923098792758538'),
    ('7475328298248178984'),
    ('7473432757239057722');

create table `user`(
    id varchar(6) primary key,
    nickname varchar(255) unique not null,
    password varchar(255) not null,
    email varchar(255) unique not null
);

create table aweme_videos(
    id varchar(36) primary key,
    aweme_id varchar(64),
    description text,
    author_uid varchar(64),
    author_sec_uid varchar(64),
    author_nickname varchar(128),
    author_avatar text,
    music_title varchar(255),
    music_author varchar(255),
    music_cover text,
    digg_count int,
    comment_count int,
    share_count int,
    video_url text,
    video_width int,
    video_height int
);

