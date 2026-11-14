# Copywriting

The copywriting skill helps an AI agent draft and edit copy that sounds like the
person or business behind it. It covers websites, product interfaces, emails,
proposals and articles, from a single button label to a complete page.

Its aim is clear, specific writing with a recognizable voice. It combines
copywriting guidance with an editing pass for common AI writing habits, while
preserving the facts, meaning and character of the original.

## When to use it

- **Draft:** Write new copy from a brief, notes or confirmed product information.
- **Rewrite:** Improve clarity, structure or tone while keeping the writer's
  angle and useful language.
- **Point edit:** Fix a headline, call to action, error message or short passage
  without rewriting everything around it.
- **Audit:** Identify the issues that most affect clarity, credibility or voice,
  with examples and practical fixes. An audit does not automatically rewrite
  the copy.

Use it when writing feels generic, inflated, repetitive or unlike your brand.
It also helps turn product features into concrete benefits and make interface
text explain what happens next.

## How it works

The agent starts with the supplied material and relevant project context:
existing copy, brand guidance, writing samples and any product marketing notes.
It works out who the reader is, what they need to understand or do, and which
facts support the message. A small edit stays small; missing information only
prompts a question when the answer would materially change the result.

Voice comes from your examples and direction. Vocabulary, formality, humor,
punctuation and sentence rhythm should fit the writer and audience. Without a
sample, the default is plain, conversational language suited to the subject.

For substantial drafts and rewrites, the agent then checks for empty openers,
vague benefits, stock slogans, forced contrasts and repeated closing lines. It
replaces those habits with useful information, while keeping intentional quirks
and technical language the reader needs. These patterns are editing signals,
not a test of whether someone used AI.

Facts remain intact throughout. The skill prohibits invented statistics,
testimonials, capabilities and personal experiences. It preserves conditions
such as eligibility, pricing terms and limits. If evidence is missing, the agent
asks or marks a specific gap, such as `[NEED: confirmed trial terms]`.

The result is finished copy in your requested format, with unresolved facts
flagged. When editing a file, the agent preserves surrounding markup, variables,
links and code unless the requested change requires otherwise.

## How to invoke it

With the skill installed and discoverable by your agent, name it in your request:

```text
Use the copywriting skill to rewrite this homepage introduction.
Keep the tone warm and direct, and preserve all factual claims.
```

In agents that support `$skill-name` invocation, use `$copywriting`:

```text
$copywriting Draft a welcome email from these onboarding notes.
The audience is new account owners. Keep it under 150 words.
```

```text
$copywriting Edit only the headline and CTA in this file.
The CTA leads to a demo booking form. Give me three alternatives.
```

```text
$copywriting Audit this proposal for generic AI phrasing.
Show the main problems and suggested fixes before rewriting.
```

The agent may also select the skill automatically for relevant writing requests.
For the best result, provide the copy or file, audience, intended action,
confirmed facts and a voice sample if you have one. State any length limit or
passages that should stay unchanged.

For project installation, use the complete `copywriting` directory as described
in the [Master Skills installation guide](https://gitlab.codepilot.com/sean-schertell/master-skills/-/blob/main/docs/installation.md).
The working instructions are in [SKILL.md](SKILL.md); the README is an overview.

## Background

Adapted from Corey Haines's Copywriting and Siqi Chen's Humanizer, with CodePilot
guidance for project voice, scope and factual accuracy. The adaptations are
bundled; neither upstream is a runtime dependency. See the
[source revisions and licenses](references/sources.md).
