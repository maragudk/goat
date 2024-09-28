-- model_types is an enum for the type of models.
create table model_types (
  v text primary key
) strict;

insert into model_types (v) values ('brain'), ('llamacpp'), ('openai'), ('anthropic');

-- models are llms.
-- They have names (how they're identified) and types (how they're communicated with),
-- as well as configuration (in JSON) which varies by type.
create table models (
  id text primary key default ('m_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  name text not null,
  type text not null references model_types (v),
  config text not null default '{}'
) strict;

create trigger models_updated_timestamp after update on models begin
  update models set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

insert into models (id, name, type, config) values
  ('m_59ff15344d498ee0db983ad592340a81', 'human', 'brain', '{}'),
  ('m_50981744360a6e19c18b053f53cc7301', 'llama3.2-1b', 'llamacpp', '{"address":"localhost:8090"}'),
  ('m_7c063c75af9370705f165f4daf700f60', 'gpt-4o', 'openai', '{"token":"123"}');

-- speakers are named models with an optional system prompt. Many speakers can use the same model.
-- Think of these as roles for models.
create table speakers (
  id text primary key default ('s_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  name text unique not null,
  modelID text not null references models (id),
  system text not null default ''
) strict;

create trigger speakers_updated_timestamp after update on speakers begin
  update speakers set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

create index speakers_name on speakers (name);

insert into speakers (id, name, modelID, system) values
  ('s_26a91be1873f385bb0631ad868bf7c85', 'me', 'm_59ff15344d498ee0db983ad592340a81', 'You do you.'),
  ('s_6a719774ed33fb3cd2b955f7eb36fc50', 'llama', 'm_50981744360a6e19c18b053f53cc7301', ''),
  ('s_afe5f56180e339ee0fa08c0a84894fab', 'penguin', 'm_50981744360a6e19c18b053f53cc7301', 'You are a weird penguin.'),
  ('s_196169d1616d094959b1f21212da6066', 'gpt', 'm_7c063c75af9370705f165f4daf700f60', '');

-- conversations have optional topics and tie turns together.
create table conversations (
  id text primary key default ('c_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  topic text not null default ''
) strict;

create trigger conversations_updated_timestamp after update on conversations begin
  update conversations set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

-- turns in a conversation, by a speaker.
create table turns (
  id text primary key default ('t_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  conversationID text not null references conversations (id) on delete cascade,
  speakerID text not null references speakers (id),
  content text not null default ''
) strict;

create trigger turns_updated_timestamp after update on turns begin
  update turns set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

create index turns_conversationID_created on turns (conversationID, created);
