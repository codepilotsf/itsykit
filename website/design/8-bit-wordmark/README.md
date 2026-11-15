# ItsyKit 8-bit wordmark

Owner-approved option 5 (Small Assembly), lettering only, renamed ItsyKit with the current site palette. The separate iw emblem is excluded. The favicon uses the Itsy substring. No font files, external type service or build step is required.

`export-logo.mjs` converts the selected specimen's lettering into crisp, rectilinear SVG geometry. Run it from `website/` with `node design/8-bit-wordmark/export-logo.mjs`. It reads the ink, pale and accent tokens from `site/css/theme.css`. The header uses navy at 132px on desktop and 110px on phones; the light asset uses the pale theme ink. `production.json` records source bounds, method and approval.

Browser Canvas produces the 32px PNG favicon and 180px touch icon from `site/images/favicon.svg`. The sharing image is rendered from `design/social-preview.html`. Update asset query hashes in both layouts after regeneration.

The selected source and the other four concepts retain their original prompts and generation provenance under `design/logo-exploration-20261114/`. Historical woven-loop assets and records remain for reference.
