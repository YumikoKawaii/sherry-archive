create table books
(
    `id`               int auto_increment primary key,
    `title`            varchar(255),
    `description`      varchar(255) default null,
    `image_url`        varchar(255) default null,
    `author_id`        int          default null,
    `publisher_id`     int          default null,
    `category_id`      int          default null,
    `publication_date` timestamp    default null,
    `created_at`       timestamp    default current_timestamp,
    `updated_at`       timestamp    default current_timestamp on update current_timestamp,
    `deleted_at`       timestamp    default null
);

create table pages
(
    `id`         int auto_increment primary key,
    `book_id`    int,
    `image_url`  varchar(255) default null,
    `index`      int          default null,
    `created_at` timestamp    default current_timestamp,
    `updated_at` timestamp    default current_timestamp on update current_timestamp,
    `deleted_at` timestamp    default null
);

create table authors
(
    `id`          int auto_increment primary key,
    `name`        varchar(255),
    `image_url`   varchar(255),
    `description` varchar(255),
    `created_at`  timestamp default current_timestamp,
    `updated_at`  timestamp default current_timestamp on update current_timestamp,
    `deleted_at`  timestamp default null
);

create table publishers
(
    `id`          int auto_increment primary key,
    `name`        varchar(255),
    `image_url`   varchar(255),
    `description` varchar(255),
    `created_at`  timestamp default current_timestamp,
    `updated_at`  timestamp default current_timestamp on update current_timestamp,
    `deleted_at`  timestamp default null
);