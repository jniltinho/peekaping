# Fake mTLS Server for PeekaPing Testing

This fake server provides mTLS endpoints for testing PeekaPing's client certificate authentication functionality.

## Quick Start

1. **Setup certificates and start server:**
   ```bash
   cd apps/fake-server
   npm run setup
   npm run mtls
   ```

2. **Get certificate contents for PeekaPing:**
   ```bash
   npm run show-certs
   ```

3. **Create a monitor in PeekaPing:**
   - URL: `https://localhost:3443/`
   - Authentication: mTLS
   - Paste the certificate contents from step 2

## Available Scripts

| Script | Description |
|--------|-------------|
| `npm run setup` | Generate certificates and prepare for testing |
| `npm run generate-certs` | Generate CA, server, and client certificates |
| `npm run show-certs` | Display certificate contents for copying |
| `npm run mtls` | Start the mTLS server on port 3443 |
| `npm run dev` | Start the regular HTTP server on port 3022 |

## Test Endpoints

The mTLS server provides several endpoints for testing:

### Main Endpoint
- **URL:** `https://localhost:3443/`
- **Description:** Returns detailed information about the request and client certificate
- **Response:** JSON with certificate details, headers, and request info

### Health Check
- **URL:** `https://localhost:3443/health`
- **Description:** Simple health check (works without client certificate)
- **Response:** `{"status": "healthy", "message": "Server is running"}`

### Certificate Required
- **URL:** `https://localhost:3443/require-cert`
- **Description:** Strictly requires a valid client certificate
- **Response:** Returns 401 if no certificate provided, otherwise certificate details

### Optional Certificate
- **URL:** `https://localhost:3443/optional-cert`
- **Description:** Works with or without client certificate
- **Response:** Indicates whether a certificate was provided

## Query Parameters

Add `?delay=X` to any endpoint to simulate network latency (X = milliseconds).

Example: `https://localhost:3443/?delay=500`

## Testing the Fix

### 1. Test Without Certificates (Should Fail)
Create a monitor in PeekaPing with:
- URL: `https://localhost:3443/require-cert`
- Authentication: None

Expected result: Monitor should fail with SSL/TLS error.

### 2. Test With Certificates (Should Work)
Create a monitor in PeekaPing with:
- URL: `https://localhost:3443/`
- Authentication: mTLS
- Use certificates from `npm run show-certs`

Expected result: Monitor should show UP status and you can see certificate details in the server logs.

### 3. Verify Form Persistence
After creating the mTLS monitor:
1. Go to edit the monitor
2. Check that the Certificate, Key, and CA fields are populated
3. This confirms the form field fix worked

## Server Logs

The mTLS server logs show detailed information:
```
GET / → 200 (150.2 ms) [cert: Test Client]
POST /api/test → 401 (50.1 ms) [no cert]
```

This helps you verify:
- Whether the client certificate was received
- The certificate subject (CN)
- Response times and status codes

## Certificate Details

The generated certificates have these properties:
- **CA:** Test CA (self-signed)
- **Server:** Valid for localhost, 127.0.0.1
- **Client:** Test Client
- **Validity:** 365 days from generation

All certificates are generated with proper extensions for mTLS testing.

## Troubleshooting

### "Certificates not found" Error
Run: `npm run generate-certs`

### "EADDRINUSE" Error (Port 3443 in use)
Either:
- Stop other services using port 3443
- Or change the port in `mtls-server.js`

### PeekaPing Shows "Certificate Required" Error
This means the fix worked! The client certificates are being sent correctly, but might be:
- Invalid/expired certificates
- Wrong CA certificate
- Certificate format issues

Check the server logs to see what certificate information was received.

## Files Generated

```
apps/fake-server/certs/
├── ca.crt          # CA certificate (for PeekaPing CA field)
├── ca.key          # CA private key
├── ca.srl          # CA serial number
├── client.crt      # Client certificate (for PeekaPing Certificate field)
├── client.key      # Client private key (for PeekaPing Key field)
├── server.crt      # Server certificate
└── server.key      # Server private key
```

## Security Note

⚠️ **These certificates are for testing only!** They are self-signed and should never be used in production.
