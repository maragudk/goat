create view conversation_history as
select conversationID, name, content from turns
  join speakers on speakers.id = speakerID
  join conversations on turns.conversationID = conversations.id
order by conversations.created, turns.created;
