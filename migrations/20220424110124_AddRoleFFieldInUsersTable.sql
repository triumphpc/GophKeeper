-- +goose Up
alter table users
    add role varchar(10) default 'user' not null;

comment on column users.role is 'Role of user';



-- +goose Down
alter table users
    drop column role;