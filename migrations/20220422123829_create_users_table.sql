-- +goose Up
create table users
(
    id serial not null
        constraint users_pk
            primary key,
    login varchar(10) not null,
    password varchar(100) not null
);

comment on table users is 'Users database';

comment on column users.id is 'Uniq identifier';

comment on column users.login is 'User login';

comment on column users.password is 'User password';

create unique index users_login_uindex
    on users (login);



-- +goose Down
drop table users;

