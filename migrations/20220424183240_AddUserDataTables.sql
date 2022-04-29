-- +goose Up
create table if not exists user_data
(
    id        serial   not null
        constraint user_data_pk
            primary key,
    type_id   smallint not null,
    entity_id integer  not null,
    user_id   integer  not null
        constraint user_data_users_id_fk
            references users
            on delete cascade
);

comment on column user_data.id is 'Unique identificator';

comment on column user_data.type_id is 'Type of data entity';

comment on column user_data.entity_id is 'Identificator of entity';

comment on column user_data.user_id is 'User identificator';

alter table user_data
    owner to postgres;


create table if not exists user_data_cards
(
    id     serial       not null
        constraint user_data_cards_pk
            primary key,
    number varchar(250) not null,
    meta   varchar(100)
);

comment on table user_data_cards is 'User data about cards';

comment on column user_data_cards.id is 'Unique identificator';

comment on column user_data_cards.number is 'Encrypted card number';

comment on column user_data_cards.meta is 'Meta information about card number';

alter table user_data_cards
    owner to postgres;

-- auto-generated definition
create table if not exists user_data_files
(
    id      serial       not null
        constraint user_data_files_pk
            primary key,
    file_id varchar(150) not null,
    meta    varchar(100),
    path    varchar(250) not null
);

comment on table user_data_files is 'Binary files storage';

comment on column user_data_files.id is 'Unique identifiers';

comment on column user_data_files.file_id is 'File identifier';

comment on column user_data_files.meta is 'Meta information about file';

comment on column user_data_files.path is 'File path';

alter table user_data_files
    owner to postgres;

create table if not exists user_data_text
(
    id         serial                              not null
        constraint user_data_text_pk
            primary key,
    text       varchar(250)                        not null,
    meta       varchar(100),
    name       varchar(50)                         not null,
    updated_at timestamp default CURRENT_TIMESTAMP not null
);

comment on table user_data_text is 'Text user data';

comment on column user_data_text.id is 'Unique identificator';

comment on column user_data_text.text is 'Text data';

comment on column user_data_text.meta is 'Meta information about text data';

comment on column user_data_text.name is 'Name of text block';

comment on column user_data_text.updated_at is 'Last time to update';

alter table user_data_text
    owner to postgres;




-- +goose Down
drop table user_data;
drop table user_data_cards;
drop table user_data_files;
drop table user_data_text;
