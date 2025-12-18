# Security Policy

## Supported Versions

We actively support the following versions of the Go Invoice Ninja SDK with security updates:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < 1.0   | :x:                |

We recommend always using the latest version of the SDK to ensure you have the most recent security updates and bug fixes.

## Reporting a Vulnerability

The Go Invoice Ninja SDK team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by emailing:

**[security@ln.software](mailto:security@ln.software)**

Alternatively, you can use GitHub's private vulnerability reporting feature:
1. Go to the [Security tab](https://github.com/AshkanYarmoradi/go-invoice-ninja/security)
2. Click "Report a vulnerability"
3. Fill out the form with details about the vulnerability

### What to Include

To help us better understand and resolve the issue, please include as much of the following information as possible:

- Type of vulnerability (e.g., authentication bypass, information disclosure, etc.)
- Full paths of source file(s) related to the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it
- Any suggested fixes or mitigation strategies

### What to Expect

- **Initial Response**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Status Updates**: We will send you regular updates about our progress (at least every 5 business days)
- **Validation**: We will work to validate and reproduce the issue
- **Resolution Timeline**: We aim to resolve critical vulnerabilities within 30 days
- **Disclosure**: Once fixed, we will coordinate with you on public disclosure timing

### Disclosure Policy

- We request that you give us reasonable time to address the issue before public disclosure
- We will credit you in our security advisory (unless you prefer to remain anonymous)
- We will create a security advisory on GitHub for confirmed vulnerabilities
- We will release a patch as soon as possible and notify users through GitHub releases

## Security Best Practices

When using the Go Invoice Ninja SDK, we recommend following these security best practices:

### API Token Management

- **Never commit API tokens** to version control
- Store tokens in environment variables or secure secret management systems
- Rotate tokens regularly
- Use separate tokens for development, staging, and production environments
- Revoke tokens immediately if they are compromised

### HTTPS/TLS

- Always use HTTPS endpoints when communicating with Invoice Ninja API
- For self-hosted instances, ensure valid TLS certificates are configured
- Never disable TLS certificate verification in production

### Webhook Security

- Always verify webhook signatures using the built-in verification methods
- Use HTTPS endpoints for webhook receivers
- Implement rate limiting for webhook endpoints
- Log and monitor webhook requests for unusual activity

### Dependencies

- Keep the SDK updated to the latest version
- Regularly audit your dependencies using `go list -m all | nancy sleuth`
- Monitor GitHub security advisories for this repository

### Code Security

```go
// ✅ Good: Using environment variables
token := os.Getenv("INVOICE_NINJA_TOKEN")
client := invoiceninja.NewClient(token)

// ❌ Bad: Hardcoded credentials
client := invoiceninja.NewClient("your-secret-token-here")
```

```go
// ✅ Good: Verify webhook signatures
if !invoiceninja.VerifyWebhookSignature(payload, signature, secret) {
    return errors.New("invalid webhook signature")
}

// ❌ Bad: Process webhooks without verification
// Don't trust incoming webhook data without verification
```

## Security Updates

Security updates will be released as follows:

1. **Critical vulnerabilities**: Immediate patch release
2. **High severity**: Patch within 7 days
3. **Medium severity**: Patch within 30 days
4. **Low severity**: Included in next regular release

Security updates will be announced through:
- GitHub Security Advisories
- GitHub Releases with security tags
- Repository README (if critical)

## Known Security Considerations

### Rate Limiting

This SDK implements client-side rate limiting, but it should not be relied upon as the sole protection mechanism. Implement your own rate limiting at the application level and monitor for abuse.

### Error Messages

Be cautious about logging error messages in production environments as they may contain sensitive information. Use structured logging and sanitize error messages before displaying them to end users.

### Data Handling

- The SDK transmits data to Invoice Ninja's API in plain text (over HTTPS)
- Ensure sensitive data is handled according to your organization's security policies
- Be aware of data residency requirements when using cloud-hosted Invoice Ninja

## Scope

This security policy applies to:
- The Go Invoice Ninja SDK source code in this repository
- Official releases and distributions

This policy does **not** cover:
- Invoice Ninja application itself (report to Invoice Ninja team)
- Third-party applications using this SDK
- User's implementation code

## Contact

For security-related questions that are not vulnerability reports, you can:
- Open a discussion on GitHub Discussions
- Email: [security@ln.software](mailto:security@ln.software)

---

Thank you for helping keep the Go Invoice Ninja SDK and its users secure!
