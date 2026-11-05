## MODIFIED Requirements

### Requirement: Persistent in-body attribution

Every Twinmate-generated message MUST retain a locale-neutral AI marker
in its body, visible on native mobile and grouped posts. Server-owned
provenance MUST identify verified generated posts. Edits MUST NOT remove
the reserved marker from already published labeled posts, even after a
provenance-storage failure; unverified labels MUST NOT confer a trusted badge.

#### Scenario: Message from a developer's local agent

- GIVEN a valid send-enabled Twinmate agent key
- WHEN the agent publishes a result as its owner
- THEN the resulting Mattermost post begins with the `🤖 Twinmate · AI-generated` label
- AND the post includes machine-readable Twinmate provenance when storage succeeds

#### Scenario: Native mobile or grouped messages

- GIVEN a Twinmate-authored message is displayed on a native mobile client
- WHEN the client cannot render Mattermost webapp-plugin components
- THEN the persistent in-body marker still identifies it as AI-generated

#### Scenario: Attempt to edit away the label

- GIVEN a post with the persisted Twinmate body marker
- WHEN its author tries to remove that marker
- THEN Mattermost rejects the edit, including when provenance KV is unavailable

#### Scenario: A permitted edit keeps its attribution

- GIVEN an ordinary human post or an AI post whose edit preserves the body label
- WHEN its author changes the post text
- THEN Twinmate permits the edit instead of rejecting it in the update hook
