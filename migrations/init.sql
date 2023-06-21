create table if not exists campus (
    id serial primary key,
    name varchar(50) not null,
    topic_id bigint not null
);

create table if not exists issues(
    id              serial          primary key ,
    tg_message_id   int             not null    ,
    key             varchar (50)    not null    ,
    link            varchar (50)    not null    ,
    summary         varchar (100)   not null    ,
    description     text            not null    ,
    campus_id       bigint          references campus(id) not null,
    reporter        varchar (50)    not null    ,
    assignee        varchar (50)                ,
    priority        varchar (50)    not null    ,
    created_date    timestamp       default CURRENT_TIMESTAMP not null
);

create table if not exists users (
    id serial primary key,
    name varchar(100) not null,
    campus_id bigint references campus(id) not null
);
