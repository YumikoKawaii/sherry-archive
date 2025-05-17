create table users
(
    id              int auto_increment primary key,
    email           varchar(50) not null,
    hashed_password text        not null,
    username        varchar(50) default null,
    department      varchar(50) default null,
    status          varchar(50) not null,
    created_at      timestamp   default current_timestamp,
    updated_at      timestamp   default current_timestamp on update current_timestamp
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
    user_id    int         not null,
    role       varchar(50) not null,
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp on update current_timestamp,
    constraint fk_users foreign key (user_id) references users (id),
    constraint fk_roles foreign key (role) references roles (name)
);

create table `policies`
(
    id             int auto_increment primary key,
    `name`         varchar(50) not null unique,
    description    text,
    effect         varchar(10) not null,
    principal_type varchar(10) not null,
    resource_id    int         not null,
    created_at     timestamp default current_timestamp,
    updated_at     timestamp default current_timestamp on update current_timestamp,
    constraint fk_permissions foreign key (resource_id) references resources (id)
);