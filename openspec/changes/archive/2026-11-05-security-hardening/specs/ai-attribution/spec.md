## MODIFIED Requirements

### Requirement: Persistent in-body attribution

Every message Twinmate publishes on behalf of its owner MUST store an
unambiguous, locale-neutral AI-generated marker in the message body, as
well as machine-readable provenance metadata. The marker MUST remain visible
in native mobile clients and when consecutive posts have their headers
hidden. A user edit MUST NOT remove the marker from server-recorded
AI-generated posts.

#### Scenario: Message from a developer's local agent

- GIVEN a valid send-enabled Twinmate agent key
- WHEN the agent publishes a result as its owner
- THEN the resulting Mattermost post begins with the `🤖 Twinmate · AI-generated` label
- AND the post includes machine-readable Twinmate provenance

#### Scenario: Native mobile or grouped messages

- GIVEN a Twinmate-authored message is displayed on a native mobile client
- WHEN the client cannot render Mattermost webapp-plugin components
- THEN the persistent in-body marker still identifies it as AI-generated

#### Scenario: Attempt to edit away the label

- GIVEN a post with server-recorded Twinmate provenance
- WHEN its author tries to remove the body marker
- THEN Mattermost rejects the edit and preserves AI attribution
