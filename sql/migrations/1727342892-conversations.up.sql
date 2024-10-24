-- model_types is an enum for the type of models.
create table model_types (
  v text primary key
) strict;

insert into model_types (v) values ('brain'), ('llamacpp'), ('openai'), ('anthropic'), ('groq'),
                                   ('huggingface'), ('fireworks'), ('google');

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
  ('m_bff0168b18e50745baed4d02a24d4b66', 'llama3.2-3b', 'llamacpp', '{"address":"localhost:8091"}'),
  ('m_32cba90058abc8856ee083461f859be4', 'llama3.1-8b', 'llamacpp', '{"address":"localhost:8092"}'),
  ('m_908680bfa1a690fe805070534cc73bed', 'llama-3.1-70b-versatile', 'groq', '{"token":"123"}'),
  ('m_7c063c75af9370705f165f4daf700f60', 'gpt-4o', 'openai', '{"token":"123"}'),
  ('m_b2ac6559f08edb63d5db48231a7d7aae', 'claude-3-5-sonnet-20240620', 'anthropic', '{"token":"123"}'),
  ('m_7f439ec1580fcc145e388f117dc0897a', 'google/gemma-2-2b-it', 'huggingface', '{"token":"123"}');

-- speakers are named models with an optional system prompt. Many speakers can use the same model.
-- Think of these as roles for models.
create table speakers (
  id text primary key default ('s_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  name text unique not null,
  modelID text not null references models (id),
  system text not null default '',
  config text not null default '{"avatar":"🤖"}'
) strict;

create trigger speakers_updated_timestamp after update on speakers begin
  update speakers set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

insert into speakers (id, name, config, modelID, system) values
  ('s_26a91be1873f385bb0631ad868bf7c85', 'me', '{"avatar":"🧑"}', 'm_59ff15344d498ee0db983ad592340a81', 'You do you.'),
  ('s_6a719774ed33fb3cd2b955f7eb36fc50', 'llama1','{"avatar":"🦙"}', 'm_50981744360a6e19c18b053f53cc7301', ''),
  ('s_7136eef88ec2628462b9b28c30327421', 'llama3', '{"avatar":"🦙"}', 'm_bff0168b18e50745baed4d02a24d4b66', ''),
  ('s_60cdf7c9203bfb3ab62d9000ea8005e1', 'llama8', '{"avatar":"🦙"}', 'm_32cba90058abc8856ee083461f859be4', ''),
  ('s_73763b5713a13b77cecf50c63066b3c5', 'llama', '{"avatar":"🦙"}', 'm_908680bfa1a690fe805070534cc73bed', ''),
  ('s_196169d1616d094959b1f21212da6066', 'gpt', '{"avatar":"🤖"}', 'm_7c063c75af9370705f165f4daf700f60', ''),
  ('s_0f2981f8f63af40eae0042502f8fbef4', 'claude', '{"avatar":"🧑"}‍🎨', 'm_b2ac6559f08edb63d5db48231a7d7aae', ''),
  ('s_b7b078c33569f1a1a7c393559a018981', 'gemma2', '{"avatar":"🦠"}', 'm_7f439ec1580fcc145e388f117dc0897a', '');

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
