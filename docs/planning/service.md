# CodePilot 2.0 — Service and Architecture Planning

Status: WORK IN PROGRESS. Agent handoff, not a final specification or implementation authorization.
Updated: October 3, 2026.
Owner: Sean, Codepilot, Inc.

Repository copy imported October 3, 2026 from the separate planning workspace.
Related: [Flea apps planning](apps.md). This copy preserves the service discussion;
later changes in the original workspace are not automatically synchronized.

## Purpose and planning approach

CodePilot is evolving from a web application agency into a human-led, AI-enabled small business growth service. Keep the CodePilot name; develop a new identity, message, and offering. Prefer “growth” and “accelerator” over “marketing.”

Goal: help owners attract more customers without adding substantial work to their day. They are good at their business but find getting noticed online overwhelming. CodePilot takes responsibility for delivery; this is not a DIY tool subscription. AI enables the economics and convenience; people set direction, review work, and care for clients.

Serve a range of small businesses based on fit, without a rigid category list. Defer unusual edge cases. Examples include plumbers, live-event venues, and professional firms.

Sean wants concise consulting, one question at a time, preferably multiple choice. Do not reopen settled decisions unnecessarily. StoryBrand copy comes LAST, after the service and architecture are sufficiently defined. Implementation details should not overwhelm the planning process.

## Commercial terms — agreed direction

- $500/month, collected at signup; no setup fee or minimum term.
- Cancel anytime. If the client does not feel they received more than $500 of value, refund that month, no questions asked. Exact contractual wording and refund mechanics remain to be written.
- Clients judge value; no requirement to disclose sensitive sales/revenue information.
- Website creation, hosting, maintenance, and reasonable changes included. No fixed revision/update count. Major new functionality, such as custom ecommerce, quoted separately.
- Free website export on departure. Alternatively, a smaller hosting-only plan; price and export packaging remain undecided.
- Domain registration/renewal charged separately at cost, approved in advance.
- Client-specific third-party tools paid separately, approved in advance, preferably in client-owned accounts.
- CodePilot delivery tools, including Cody, covered by the subscription.
- Approximately four HUMAN hours per client monthly is an internal ongoing target, not a client-facing cap. Reasonable overages are handled and tracked. Month-one setup has a separate budget, not yet quantified.
- Initial team: Sean, possibly one additional person later. Shared work queue rather than dedicated account managers.
- Pilot client count undecided; Sean explicitly deferred it.

## First month — standard setup checklist

Use one first-month setup commitment, not multiple deadlines. The earlier two-week website deadline was superseded. Completion depends on access, information, and approvals; precise treatment of delays remains open.

1. Show a standard welcome video from Sean introducing the service and Cody.
2. Cody conducts onboarding interviews, up to 30 minutes total, split into shorter sessions if useful.
3. Gather business details, goals, services, audience, voice, brand assets, existing content, and staff roles.
4. Audit the existing website: useful pages/content, important URLs, forms, booking, payments, ecommerce, memberships, and integrations.
5. Flag functionality that does not readily move to Flea. Retain/link/embed existing tools when practical, or agree a replacement or separately priced integration. Client approves exceptions before migration.
6. Identify registrar, DNS provider, hosting, and email provider; assist with recovering account access and granting CodePilot access. Do not expect owners to understand DNS instructions.
7. Cody guides through an app checklist. Sean personally helps if necessary; included onboarding assistance. Screen sharing is a possible future tool, not an MVP requirement.
8. Propose a growth plan: primary CTA, channels, themes, and priorities. Agent preparation is acceptable; Sean reviews, chooses direction, and records a personalized plan video.
9. Client explicitly approves the plan (“Looks good”) before production.
10. Build/migrate the Flea website, bringing over useful content and preserving important URLs or adding redirects. Flag unusually large sites in the audit.
11. Configure direct inquiry delivery to the client, analytics, relevant profiles, review tools, and publishing connections.
12. Establish content guidelines, review rules, interviews, photo prompts, and reserve content.
13. Client explicitly approves website preview before launch.
14. Manage DNS/SSL/cutover; preserve and verify email and other existing services. Domain registration/renewal management varies by client; client remains owner and renewal responsibility is recorded.
15. Keep old-site backup and rollback through verification, then help retire old hosting with explicit client approval before cancellation.

Internal readiness checklist flags missing items but allows independent work to proceed. Launch requires its actual prerequisites.

## Ongoing delivery — starts in month two

- Two CORE social posts weekly, adapted for selected platforms; not two unique posts per platform.
- One substantive blog article monthly, grounded in client expertise and search opportunities.
- Website content updates and reasonable changes without fixed quantity limits.
- Foundational and ongoing SEO/AEO: useful answers, accurate business facts, relevant structured data, service/location content, indexing, metadata, redirects, and iterative improvements. No promise of ranking or inclusion in AI answers.
- Google Business Profile setup and ongoing updates where relevant.
- Review tools: link, QR code, and request templates that CLIENTS send.
- Basic editing of client clips: trimming, captions, basic audio cleanup, simple branding. Video counts as one weekly post. No video quota. Filming, elaborate animation, and complex production separate.
- Branded graphics, licensed stock, and selective AI illustrations when suitable. Never fabricate actual people, projects, or customer results.
- Two weeks of approved evergreen reserve content: four core posts, replenished as used. Continue publishing if interviews/assets are missed.
- Uptime monitoring, automated backups, and inquiry-form checks.
- Monthly internal performance review. Agent prepares notes; Sean verifies findings, chooses next steps, and records a short client video. Next steps come from Sean in the video, not a separate Cody-generated action request.
- Dashboard shows visits and measurable CTA actions, plus a linked list of completed work. Distinguish phone-link clicks from calls and booking-link clicks from completed bookings.

Initial exclusions: customer conversations, comments/DM handling, sales pipeline management, email newsletters, broad custom operational work. Clients handle all customer communication. Inquiry handling initially means direct delivery and links to existing booking tools.

Paid ads: DEFERRED until after pilot. Longer-term intent is management included when appropriate, client-funded ad spend; scope unvalidated.

Future Flea mini-app menu: blog, events, booking, and other basic operational tools. These are a roadmap, not all current Flea features.

## Content review and publishing

- Agents automatically prepare weekly batches using interviews, uploads, and the content plan.
- Human review first. Sean can edit directly OR ask the agent for revisions.
- Client receives one weekly batch and two business days to request changes, agreed during onboarding.
- Articles use the same review-window process.
- Scheduled publication is automatic after human approval and client window, unless changes requested.
- Client revisions go through Cody via text/voice on the draft screen; requests remain linked to drafts. Affected publication pauses until resolved.
- Client approval can eventually be delegated to CodePilot. Initially the agreed review-window rule applies; exact delegation settings remain to define.
- Initial website launch and growth plan require EXPLICIT client approval.
- Routine website changes get internal human review and may publish without client approval; substantial or unclear changes get client preview/confirmation.

## Cody and client app

Hotwire Native for iPhone AND Android at launch, with shared Rails screens and native voice/push integrations. Browser portal supports website reviews and desktop use.

Cody is one consistent CLIENT-FACING identity, not one model doing everything. Voice interviewer, content agent, and coding agent can be distinct systems sharing Rails client records.

App scope:
- Onboarding, access/assets checklist, progress.
- Individual ongoing conversations with Cody; text, attachments, and voice messages.
- In-app voice interviews, photos/video uploads, content batches and revisions.
- Website requests and separate status list, alongside conversation.
- Dashboard, completed work, monthly videos.
- Ask Cody for a human OR explicit “Contact the team.” Human responses clearly labeled in the same conversation.
- Human acknowledgment within one business day; not a promise to complete every request in that time.
- Clients do NOT need Flea admin editing access initially. CodePilot handles edits.

Owner plus invited staff:
- Staff contribute assets and can interview with Cody.
- Owner controls content approvals and sees business requests/content.
- Staff can directly submit website requests without prior owner approval; CodePilot reviews them.
- Each person has their own conversation. Relevant material feeds shared business memory; attribution retained.
- Owner selects interview-eligible staff; Cody coordinates topics and invitations.

## Interview behavior

- ElevenLabs Agents is the provisional voice provider. Quality first, with cost discipline; test before commitment.
- Interviews inside app, not regular telephone calls. Recurring invitation schedule plus on-demand start.
- Ongoing maximum: 20 minutes per BUSINESS per week, shared across owner/staff. Maximum, not duration target. A useful social interview may take 5–10 minutes; deeper articles may take longer.
- Initial onboarding allowance is separate: up to 30 minutes.
- Cody remembers business context and prior conversations, follows up, and actively uncovers specific stories and expertise.
- Start with purpose, estimated duration, and readiness check: “This will probably take around five minutes, okay?” Ask before extending beyond estimate, respecting maximum.
- Push invitation includes topic and estimate.
- Start now; remind in 10 minutes, 1 hour, or tomorrow.
- Unanswered invitation: retry after four hours, then once next business day, respecting notification hours and explicit snoozes.
- Reminders per PERSON; another employee’s interview does not cancel their invitation because they may have different information.
- Save audio and transcripts. Recording consent; separate permission before publishing a person’s voice.
- Finish with brief recap of key facts and proposed topics, allowing correction.
- Retention duration, consent implementation, quiet-hour defaults, and allocation when weekly time is exhausted remain open.

Pricing observed October 3: ElevenLabs standard additional minutes $0.08/minute, LLM extra; 20 minutes voice portion $1.60. Earlier $10/client/month budget was a planning target, not a verified all-in quote. A cheaper custom transcription/text/TTS chain was discussed; Sean favored paying more for ElevenLabs experience. Validate actual cost, mobile SDK behavior, interruptions, and reconnection.

## Architecture

| Component | Responsibility |
| --- | --- |
| Rails | Client accounts, permissions, business memory, files, drafts, checklists, task state, approvals, schedules, audit/activity history |
| Solid Queue | Agent jobs, scheduled publishing, retries, failure escalation |
| ElevenLabs/Cody | Live voice interviews and clarification; structured handoff to Rails |
| Direct model APIs | Routine article/caption/summary generation from Rails, rather than a general agent platform |
| Separate coding runtime | Isolated client checkout, Flea guidance, code changes, PR/preview, results returned to Rails; runtime undecided |
| Internal dashboard | Shared queue, context, drafts/previews, manual editing, agent revision instructions, approvals, deployment/publishing |
| Flea sites | One repository and site per client; live plus persistent private staging |
| Upload-Post | Provisional API publishing service for Instagram, Facebook, LinkedIn, GBP |
| Self-hosted Umami | Collect analytics; Rails fetches API data and renders CodePilot dashboard |

Agent task statuses: queued, working, needs information, ready for review, awaiting client, completed, failed. Common vocabulary, task-specific details. Temporary failures retry; unresolved failures alert Sean. Prevent duplicate posts/deployments, especially ambiguous timeouts.

Every content job gets a standard brief checklist:
- Business facts, services, locations, audience.
- Growth goal and primary CTA.
- Voice/tone and approved examples.
- Current offers/events and dates.
- Relevant attributed interview excerpts and approved assets.
- Recent content to avoid repetition.
- Platform, format, purpose.
- Claims needing verification and topics to avoid.

Draft with flagged missing facts rather than block all drafting. Publication blocked until essential facts are confirmed. Exact brief schema and source/approval tracking remain to design.

## Flea: current behavior versus planned enhancements

Docs: https://fleacms.com/docs
Local framework source: /Users/sean/Projects/flea
Read its AGENTS.md before implementation. No framework changes have been made by this planning conversation.

Verified current behavior from internal/admin/content.go and internal/snips/store.go:
- Templates/assets and editable JSON are separate files; no content database required.
- Normal deployment preserves live snip values by ID.
- New referenced IDs use initial local values when not present remotely (including checking archived values first).
- Removed values are archived, not automatically deleted.
- Thus adding `<p snip>Lorem ipsum</p>` locally can introduce a new editable section without replacing existing live edits.
- Explicit snip synchronization uses pull baseline/conflict handling; normal code deployment is distinct.
- Branches can have separate deployment destinations.
- Existing theme synchronization preserves remote theme during normal deployment.
- Backups and code rollback exist. Content revision-history UI, blog publishing API, and scheduled content workflow are not documented current features.

Direction discussed, NOT implemented:
- Remote/live is authoritative for published content.
- Automatically pull fresh remote content when local dev starts; preserve local work and handle offline/unconfigured behavior safely (details open).
- Sean is inclined to REMOVE remote Theme editor; CSS/theme become developer-managed. This was supported as simplification, not implemented.
- Always permit manual mode: direct live content edits, or local template/code edits with preview/deploy, alongside agent automation.
- Proposed full content API/draft/revision system was explored, then DEFERRED in favor of simpler existing Flea workflow. Do not treat it as a committed prerequisite.

Preview/release requirements:
- One persistent PRIVATE staging site per client, one website change under review at a time initially.
- Start each new preview from fresh live content, then proposed changes; do not refresh over pending work.
- Review UI in CodePilot portal: summary, preview link, Approve, Request changes. Rails owns approvals.
- Successful publication leaves review page saying “All latest changes have been published. Nothing to review right now.” Failed publication stays pending and alerts team.
- Preserve unrelated live content changes. Overlapping edits pause for reconciliation.
- PR/branch and preview for code changes; approved deployment follows human review (client approval where required).
- Promotion/conflict mechanics and content-publishing path still need design; existing merge behavior alone does not prove all of this is implemented.

## Publishing and analytics decisions

Social platform scope: Instagram, Facebook, LinkedIn, Google Business Profile. Actual channels selected per business. No promise every business uses all four.

Direct platform integrations were preferred if easy, then research favored an intermediary to avoid approval/token/media complexity. Rails retains approvals and schedules; Solid Queue calls provider API.

Upload-Post provisional, not subscribed/tested:
- Supports four desired channels; account linking, token refresh, publishing, scheduling, analytics advertised.
- Basic advertised $24 month-to-month, five client profiles; $16/month equivalent annual. Professional annual $400, 25 profiles, white-label integration. Sean dislikes annual upfront outlay; prefer free testing and monthly payment.
- Verify exact API/embedded linking requirements, white-label tier, publication counts, formats, callbacks, reliability, and terms before purchase.
- Other researched options: Post for Me $10/month/1,000 successful posts, unlimited accounts but GBP not listed; Outstand $19/month/3,000 posts but own GBP credentials and agency/resale constraints. Not selected.
- Price observations are dated, not guaranteed; recheck when implementing.

Umami direction:
- Self-host one shared instance; website tracker sends events, Rails reads API. Clients never see Umami UI, so no built-in white-label feature needed.
- Source verification by another agent found website-count cap only under CLOUD_MODE in current master website creation route. Self-hosted with CLOUD_MODE unset has no application website-count cap; infrastructure remains limiting factor. This can be checked off.
- MIT-licensed source; preserve license notices.
- Cloud white-label controls are not assumed present in OSS. Earlier assertion that cloud/self-hosted feature sets are identical was overly broad.
- Hosting estimate $10–20/month shared initially if separate server, not validated capacity quote; potentially no incremental server cost if spare existing capacity.
- API auth, tenant isolation in Rails, event taxonomy, retention, and hosting sizing remain to implement.

## Brand/StoryBrand direction — not final copy

- Keep CodePilot. New visual identity not developed.
- Business owner is hero; CodePilot is guide.
- Main problem: too busy running business to manage online growth; getting noticed feels overwhelming.
- Desired success: more customers without adding more work.
- Lead with outcome; human-led/AI explanation in supporting copy.
- Credibility: experience, client relationships, testimonials; do not invent attributable growth results.
- Main CTA: “Get started,” direct signup, not application or mandatory sales meeting. Explain scope before payment.
- Three-step plan chosen: Sign up → Share your expertise → We run your growth engine.
- Draft lines discussed (“Grow your business without adding more to your plate”) are exploratory, not approved finished messaging.
- Be clear about limited client participation, month-one setup, ongoing deliverables, separate costs, guarantee, and export/hosting choices.

## Next work and open decisions

Recommended next planning topic: blog creation and scheduled publishing, because monthly article is committed but Flea implementation is not yet defined.

Other open work:
1. Coding runtime, isolation, credentials, repository host, deployment workers, and server topology.
2. Flea dev auto-pull, theme-editor removal decision, staging refresh/promotion, overlap detection, rollback, and safe introduction of new content structures.
3. Image/video generation and processing tools; storage for interview audio/client media (Flea currently has small video limits and no transcoding).
4. Upload-Post validation and plan choice; GBP profile-detail management may need separate integration from post publishing.
5. Umami setup, API reports, CTA event schema, bot filtering, verified versus click-only conversions.
6. Consent, retention, file permissions, staff visibility, secure delegated account access.
7. Notification settings and interview cap edge cases.
8. Setup time, ongoing unit economics, pilot success criteria; pilot size intentionally deferred.
9. Hosting-only price, free export scope, guarantee wording/process, exceptions when onboarding stalls.
10. Final client-facing service checklist, then StoryBrand copy and brand identity.

Do not treat this document as approval to build, deploy, contact clients, purchase tools, or change Flea. It records planning decisions and distinguishes intended behavior from verified implementation.

## Reference links

- Flea: https://fleacms.com/docs
- ElevenLabs Agents pricing: https://elevenlabs.io/pricing/agents
- ElevenLabs Swift SDK: https://elevenlabs.io/docs/eleven-agents/libraries/swift
- Upload-Post: https://www.upload-post.com/
- Upload-Post docs: https://docs.upload-post.com/
- Umami docs: https://docs.umami.is/docs
- Umami source: https://github.com/umami-software/umami
- Website cap source: https://github.com/umami-software/umami/blob/master/src/app/api/websites/route.ts
- Apple companion app rule: https://developer.apple.com/app-store/review/guidelines/#other-purchase-methods

Apple distribution: free existing-client companion app, website billing, no in-app purchasing was discussed as a plausible model under 3.1.3(f), not guaranteed App Review approval. Human service involvement alone does not establish an exemption. Reverify current regional rules before submission.
