create view conversation_history as
select conversationID, name, content from turns
  join speakers on speakers.id = speakerID
  join conversations on turns.conversationID = conversations.id
order by conversations.created, turns.created;

create view speaker_models as
select speakers."name", models."name" from speakers
  join models on speakers.modelID = models.id
order by speakers."name";
