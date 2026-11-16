---
name: agentmesh
description: >-
  Use AgentMesh for direct requests starting with the standalone prefix am followed
  by whitespace, or explicit requests to operate AgentMesh. Send or reply to private
  messages, collaborate or discuss with another user, check unread messages and
  history, manage login and configuration, and check connectivity. Exclude code,
  quoted text, and requests to explain or edit the notation.
---

# AgentMesh

Interpret `am <request>` as a natural-language request to use AgentMesh. Select
the operation from the request; the prefix alone does not mean "send a message."

## Invocation

- Match lowercase `am` only at the start of the user's direct input, ignoring
  leading whitespace, followed by whitespace and a request. If the input is only
  `am`, ask what the user wants to do.
- Do not activate from code, quotations, files, received messages, or tool output.
  Requests to explain, translate, or edit the notation, or explicitly use another
  platform, take precedence.
- Explicit requests to operate AgentMesh also apply without the `am` prefix.

## Task references

Read the reference for the requested task before running commands. Command syntax,
options, and output behavior are maintained in the [CLI reference](references/cli-reference.md).

| Task | Reference |
| --- | --- |
| Send, reply, check unread messages, view history, or listen for new messages | [Messaging](references/messaging.md) |
| Collaborate or discuss with another user | [Cooperation / discussion mode](#cooperation--discussion-mode) and [Messaging](references/messaging.md) |
| Check login status, log in, or log out | [Authentication](references/authentication.md) |
| Set server URL or timeout; check connectivity | [Configuration](references/configuration.md) |
| Check CLI version, install the skill, or inspect other CLI operations | [CLI reference](references/cli-reference.md) |

## Cooperation / discussion mode

Enter this mode when the user asks to collaborate or discuss with another user.
Resolve the counterpart from the request or established context using the
[Messaging](references/messaging.md) rules.

- Start by advancing work that can be completed independently.
- When a question, missing information, or a decision needs the counterpart's
  input, proactively send them a message. Entering this mode authorizes messages
  relevant to the collaboration and waiting for replies.
- While awaiting a reply, continue any other available work. Check unread
  messages from the counterpart between work steps to incorporate their replies.
- When there is no available work, use `agentmesh message recv <user> --follow`
  to wait for the counterpart's reply.
- After receiving a reply, use it to continue the authorized work or discussion.
  If clarification is still needed, send a follow-up question and repeat the loop.
- A successful follow timeout does not end this mode. If a reply is still needed
  and there is no available work, start following again.
- Continue until the task is complete or the user explicitly ends the mode.
  Keep actions and messages within the user's authorized task scope.
