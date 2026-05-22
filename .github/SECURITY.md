# Security Policy

## Supported Versions

Only the latest released version of MCPSync receives security fixes.

| Version | Supported |
| ------- | --------- |
| latest  | ✅        |
| older   | ❌        |

## Reporting a Vulnerability

**Please do not open a public GitHub issue for security vulnerabilities.**

Report privately via GitHub's Security Advisory feature:
**https://github.com/anush-data-portfolio/MCPSync/security/advisories/new**

We will acknowledge your report within **5 business days** and aim to
release a fix within **30 days** for confirmed critical issues.

## Supply-Chain Integrity

MCPSync distributes pre-built binaries via npm and GitHub Releases.
Every release includes a `checksums.txt` with SHA-256 digests.
The `postinstall.js` script verifies the downloaded binary against
this file before installing. To verify manually:

```sh
sha256sum mcpsync-<platform>
# compare against checksums.txt from the same release
```
