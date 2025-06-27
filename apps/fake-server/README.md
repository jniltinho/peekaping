# Fake Server for PeekaPing Testing

This fake server provides both mTLS endpoints and redirect testing functionality for comprehensive PeekaPing monitor testing.

## Quick Start

### HTTP Server with Redirect Testing
```bash
cd apps/fake-server
npm run dev
```
Visit http://localhost:3022/help for redirect testing endpoints.

### mTLS Server
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
| `npm run dev` | Start the HTTP server on port 3022 (includes redirect testing) |
| `npm run setup` | Generate certificates and prepare for mTLS testing |
| `npm run generate-certs` | Generate CA, server, and client certificates |
| `npm run show-certs` | Display certificate contents for copying |
| `npm run mtls` | Start the mTLS server on port 3443 |

## HTTP Server Endpoints (Port 3022)

### Basic Endpoints
- **URL:** `http://localhost:3022/`
- **Methods:** GET, POST, PUT, DELETE, PATCH
- **Description:** Returns basic response with optional delay
- **Query Parameters:**
  - `delay`: Delay in ms (default: 100)
  - `delayRandom`: Random additional delay in ms (default: 20)

### Redirect Testing Endpoints

#### Fixed Number of Redirects
- **URL:** `http://localhost:3022/redirect`
- **Methods:** GET, POST, PUT, DELETE, PATCH
- **Description:** Redirects a specified number of times before returning success
- **Query Parameters:**
  - `redirects`: Number of redirects (default: 3)
  - `status`: HTTP redirect status code - 301, 302, 303, 307, 308 (default: 302)
  - `delay`: Delay between redirects in ms (default: 0)

**Examples:**
```bash
# Test 5 redirects (should work with max_redirects >= 5)
http://localhost:3022/redirect?redirects=5

# Test 15 redirects with 301 status (should fail with max_redirects < 15)
http://localhost:3022/redirect?redirects=15&status=301

# Test redirects with delay
http://localhost:3022/redirect?redirects=3&delay=500
```

#### Infinite Redirects
- **URL:** `http://localhost:3022/infinite-redirect`
- **Methods:** GET, POST
- **Description:** Continuously redirects (never stops)
- **Query Parameters:**
  - `status`: HTTP redirect status code (default: 302)
  - `delay`: Delay between redirects in ms (default: 0)

**Examples:**
```bash
# Test infinite redirects (should fail when max_redirects limit is reached)
http://localhost:3022/infinite-redirect

# Test infinite redirects with delay
http://localhost:3022/infinite-redirect?delay=100&status=307
```

#### Circular Redirects
- **URL:** `http://localhost:3022/circular-redirect-a` or `http://localhost:3022/circular-redirect-b`
- **Description:** Creates A→B→A redirect loop
- **Query Parameters:**
  - `status`: HTTP redirect status code (default: 302)
  - `delay`: Delay between redirects in ms (default: 0)

**Examples:**
```bash
# Test circular redirects (should fail when max_redirects limit is reached)
http://localhost:3022/circular-redirect-a
```

### Help Endpoint
- **URL:** `http://localhost:3022/help`
- **Description:** Returns detailed documentation of all endpoints

## Testing Max Redirects Feature

### Test Scenarios

#### 1. Test Within Redirect Limit
Create a monitor in PeekaPing with:
- URL: `http://localhost:3022/redirect?redirects=5`
- Max Redirects: 10

**Expected:** Monitor should show UP status after following 5 redirects.

#### 2. Test Exceeding Redirect Limit
Create a monitor in PeekaPing with:
- URL: `http://localhost:3022/redirect?redirects=15`
- Max Redirects: 10

**Expected:** Monitor should show DOWN status due to exceeding redirect limit.

#### 3. Test Infinite Redirects
Create a monitor in PeekaPing with:
- URL: `http://localhost:3022/infinite-redirect`
- Max Redirects: 5

**Expected:** Monitor should show DOWN status after 5 redirects.

#### 4. Test Zero Redirects (Disabled)
Create a monitor in PeekaPing with:
- URL: `http://localhost:3022/redirect?redirects=1`
- Max Redirects: 0

**Expected:** Monitor should show DOWN status (redirects disabled).

#### 5. Test Different Redirect Types
Test various HTTP redirect status codes:
```bash
# Temporary redirect
http://localhost:3022/redirect?redirects=3&status=302

# Permanent redirect
http://localhost:3022/redirect?redirects=3&status=301

# See Other
http://localhost:3022/redirect?redirects=3&status=303

# Temporary redirect (preserve method)
http://localhost:3022/redirect?redirects=3&status=307

# Permanent redirect (preserve method)
http://localhost:3022/redirect?redirects=3&status=308
```

## mTLS Test Endpoints (Port 3443)

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

Both servers support `?delay=X` parameter to simulate network latency (X = milliseconds).

## Server Logs

Both servers provide detailed logging:

**HTTP Server:**
```
GET /redirect?redirects=5 → 302 (12.3 ms)
Redirecting (1/5) to: http://localhost:3022/redirect?redirects=5&current=1
GET /redirect?redirects=5&current=1 → 302 (8.1 ms)
Redirecting (2/5) to: http://localhost:3022/redirect?redirects=5&current=2
...
GET /redirect?redirects=5&current=5 → 200 (5.2 ms)
```

**mTLS Server:**
```
GET / → 200 (150.2 ms) [cert: Test Client]
POST /api/test → 401 (50.1 ms) [no cert]
```

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

### "EADDRINUSE" Error (Port in use)
Either:
- Stop other services using the port
- Or change the port in the respective server file

### PeekaPing Shows "Certificate Required" Error
This means the fix worked! The client certificates are being sent correctly, but might be:
- Invalid/expired certificates
- Wrong CA certificate
- Certificate format issues

Check the server logs to see what certificate information was received.

### Max Redirects Testing Issues
- Ensure your PeekaPing monitor's max_redirects setting matches your test scenario
- Check the fake server logs to see the redirect chain
- Use the `/help` endpoint for quick reference

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
