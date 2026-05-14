# Security Policy

## Supported Versions

The following versions of Raptor are currently being supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| v0.1.x  | :white_check_mark: |
| < v0.1  | :x:                |

## Reporting a Vulnerability

We take the security of Raptor seriously. If you believe you have found a security vulnerability, please do not report it through a public GitHub issue.

Instead, please send an email to **security@example.com** (replace with your actual email).

### What to include in your report

- A description of the vulnerability and its potential impact.
- Step-by-step instructions to reproduce the issue.
- Any relevant configuration details or environment information.

### What to expect from us

- We will acknowledge receipt of your report within 48 hours.
- We will keep you informed of our progress as we investigate and remediate the issue.
- Once the issue is resolved, we will provide a public disclosure of the vulnerability, including credit to you for your responsible disclosure.

## Security Guarantees

Raptor is designed to be **safe by default**:
- **No shell access**: Raptor does not expose a raw shell to the AI.
- **Input Validation**: All version strings and service names are character-whitelisted.
- **SSH Isolation**: Every command is executed via a controlled SSH session with no shared state between tool calls.
