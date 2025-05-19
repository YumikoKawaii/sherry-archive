create table tracking_ids
(
    id           varchar(36) primary key,
    `created_at` timestamp default current_timestamp,
    `updated_at` timestamp default current_timestamp on update current_timestamp,
    `deleted_at` timestamp default null
);