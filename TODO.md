# TODO

## internal/adapters/configstore
- The config is stored as JSON, but it's not versioned. 

## internal/adapters/datastore
- Put the DB in .config/data

## chat CRUD plan
- Read: expose conversation summaries in a dedicated chat side panel list, load on startup, and support selecting a historical conversation to hydrate full message history.
- Update: add rename and archive actions for conversations, persist edits through dedicated bridge methods, and optimistically update the frontend signal store.
- Delete: add hard-delete with explicit confirmation, remove conversations from signal state, and clear active conversation when the deleted chat is active.

## internal/adapters/providers
- What does cache-sqlite.go do? How does it work with registry.go?
- How are errors.go, http_client.go, provider specific?

## Architecture
- Keep the architecture, but align it by feature.

## Chat
- Clicking on 'New Chat' shouldn't launch a new chat if no messages have been sent. 
