create table sessions (
  token text primary key,
  data blob not null,
  expiry real not null
) strict;

create index sessions_expiry_idx on sessions (expiry);
