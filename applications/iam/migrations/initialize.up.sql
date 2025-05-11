create table users
(
    id         int auto_increment primary key,
    uuid       varchar(36) not null unique,
    username   varchar(50) not null,
    email      varchar(50) not null,
    department varchar(50) not null,
    status     varchar(50) not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp
);

create table `groups`
(
    id          int auto_increment primary key,
    `name`      varchar(50) not null unique,
    description text,
    created_at  timestamp default current_timestamp,
    updated_at  timestamp default current_timestamp on update current_timestamp
);

create table user_groups
(
    id         int auto_increment primary key,
    user_id    int not null,
    group_id   int not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp
);

create table resources
(
    id          int auto_increment primary key,
    `name`      varchar(255) not null,
    path        varchar(50)  not null,
    description text,
    created_at  timestamp default current_timestamp,
    updated_at  timestamp default current_timestamp on update current_timestamp
);

create table `permissions`
(
    id          int auto_increment primary key,
    `name`      varchar(50) not null,
    description text,
    created_at  timestamp default current_timestamp,
    updated_at  timestamp default current_timestamp on update current_timestamp
);

create table permission_resources
(
    id            int auto_increment primary key,
    permission_id int         not null,
    resource_id   int         not null,
    action        varchar(50) not null, # GET - POST - PUT - DELETE
    created_at    timestamp default current_timestamp,
    updated_at    timestamp default current_timestamp on update current_timestamp,
    constraint fk_resources foreign key (resource_id) references resources (id),
    constraint fk_permissions foreign key (permission_id) references permissions (id)
);

create table `roles`
(
    id          int auto_increment primary key,
    `name`      varchar(50) not null,
    description text,
    created_at  timestamp default current_timestamp,
    updated_at  timestamp default current_timestamp on update current_timestamp
);

create table user_roles
(
    id         int auto_increment primary key,
    user_id    int not null,
    role_id    int not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp
);

create table group_roles
(
    id         int auto_increment primary key,
    group_id   int not null,
    role_id    int not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp
);

create table `policies`
(
    id             int auto_increment primary key,
    `name`         varchar(50) not null unique,
    description    text,
    effect         varchar(10) not null,
    principal_type varchar(10) not null,
    principal_id   int         not null,
    permission_id  int         not null,
    created_at     timestamp default current_timestamp,
    updated_at     timestamp default current_timestamp on update current_timestamp,
    constraint fk_permissions foreign key (permission_id) references permissions (id)
);