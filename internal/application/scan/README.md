# Scan Feature - Documentation

## Overview

The **Scan** feature is a complete barcode scanning system with device information tracking. It follows Domain-Driven Design (DDD) principles and integrates seamlessly with the existing application architecture.

## Architecture

### Layers

The Scan domain is organized into six layers following DDD principles:

```
Request → Handler → Service → Repository → Entity → Database
           ↓
         Logging & Error Handling
```

#### 1. **Handler Layer** (`handler/scan.handler.go`)
- Manages HTTP requests and responses
- Validates input and error handling
- Extracts user context from middleware
- Registers all routes
- **Endpoints:**
  - `POST /api/v1/scans` - Create scan (201)
  - `GET /api/v1/scans` - List scans (200)
  - `GET /api/v1/scans/:id` - Get single scan (200)
  - `GET /api/v1/users/:user_id/scans` - Get user scans (200)
  - `PUT /api/v1/scans/:id` - Update scan status (200)
  - `DELETE /api/v1/scans/:id` - Delete scan (200)

#### 2. **Service Layer** (`service/scan.service.go`)
- Contains business logic
- Validates DTOs before persistence
- Orchestrates repository operations
- Converts entities to response DTOs
- **Methods:**
  - `CreateScan()` - Create and validate scan
  - `GetScanByID()` - Retrieve single scan
  - `GetScans()` - List all scans with pagination
  - `GetUserScans()` - Get user-specific scans
  - `UpdateScanStatus()` - Update scan status
  - `DeleteScan()` - Soft delete scan

#### 3. **Repository Layer** (`repository/scan.repo.go`)
- Handles all database operations
- Provides data access abstraction
- Uses GORM ORM
- **Methods:**
  - `Create()` - Insert scan
  - `GetByID()` - Retrieve by ID
  - `GetByTransactionID()` - Lookup by transaction
  - `GetAll()` - List with filters and pagination
  - `GetByUserID()` - User-specific queries
  - `Update()` - Update scan
  - `Delete()` - Soft delete

#### 4. **Entity Layer** (`entity/scan.entity.go`)
- Defines domain model and business rules
- GORM database model
- Status validation
- JSON marshaling for device info
- **Fields:**
  - `ID`, `UserID`, `Barcode`, `Timestamp`, `TransactionID`
  - `PIN`, `Photo`, `DeviceInfo`, `PhotoSize`
  - `Status`, `CreatedAt`, `UpdatedAt`, `DeletedAt`

#### 5. **DTO Layer** (`dto/scan.dto.go`)
- Request/response contracts
- Input validation via binding tags
- Data transformation
- **Structures:**
  - `CreateScanRequest` - Create scan input
  - `UpdateScanRequest` - Update scan input
  - `ScanResponse` - Single scan response
  - `ScanListResponse` - Paginated list response
  - `DeviceInfo` - Device metadata

#### 6. **Module Layer** (`module.go`)
- Dependency injection configuration using Uber FX
- Provides providers for DI container
- Wires dependencies

## Database Schema

Table: `scans`

| Column | Type | Properties | Purpose |
|--------|------|-----------|---------|
| `id` | BIGINT UNSIGNED | PK, AUTO_INCREMENT | Unique scan identifier |
| `user_id` | BIGINT UNSIGNED | NOT NULL, FK | Associated user |
| `barcode` | VARCHAR(255) | NOT NULL | Barcode data |
| `timestamp` | BIGINT | NOT NULL | Scan timestamp (Unix ms) |
| `transaction_id` | VARCHAR(255) | NOT NULL, UNIQUE | Unique transaction reference |
| `pin` | VARCHAR(255) | NOT NULL | PIN code |
| `photo` | LONGTEXT | NOT NULL | Base64 encoded photo |
| `device_info` | JSON | Nullable | Device metadata |
| `photo_size` | VARCHAR(50) | Nullable | Photo file size |
| `status` | VARCHAR(20) | DEFAULT 'pending' | Scan status (enum) |
| `created_at` | TIMESTAMP | DEFAULT NOW | Creation timestamp |
| `updated_at` | TIMESTAMP | DEFAULT NOW | Last update timestamp |
| `deleted_at` | TIMESTAMP | Nullable | Soft delete marker |

**Indexes:** user_id, barcode, transaction_id, status, created_at, deleted_at

**Status Values:** `pending`, `completed`, `failed`

## API Endpoints

### Create Scan
```bash
POST /api/v1/scans
Content-Type: application/json

{
  "barcode": "BC123456789",
  "timestamp": 1234567890000,
  "transaction_id": "trx-uuid-001",
  "pin": "123456",
  "photo": "data:image/jpeg;base64,/9j/4AAQSkZJRg...",
  "device": {
    "user_agent": "Mozilla/5.0 (Linux; Android 10)",
    "platform": "Linux",
    "language": "en-US",
    "device_type": "mobile",
    "browser": "Chrome"
  },
  "photo_size": "45 KB"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "user_id": 1,
    "barcode": "BC123456789",
    "timestamp": 1234567890000,
    "transaction_id": "trx-uuid-001",
    "pin": "123456",
    "photo": "data:image/jpeg;base64,...",
    "device": {
      "user_agent": "Mozilla/5.0",
      "platform": "Linux",
      "language": "en-US",
      "device_type": "mobile",
      "browser": "Chrome"
    },
    "photo_size": "45 KB",
    "status": "pending",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Get All Scans
```bash
GET /api/v1/scans?page=1&page_size=10&status=pending
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "barcode": "BC123456789",
      "status": "pending",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 10
}
```

### Get Single Scan
```bash
GET /api/v1/scans/1
```

**Response (200 OK):** Same as create response

### Get User Scans
```bash
GET /api/v1/users/1/scans?page=1&page_size=10
```

**Response (200 OK):** Paginated list of user's scans

### Update Scan Status
```bash
PUT /api/v1/scans/1
Content-Type: application/json

{
  "status": "completed"
}
```

**Response (200 OK):** Updated scan object

### Delete Scan
```bash
DELETE /api/v1/scans/1
```

**Response (200 OK):** Confirmation message

## Error Responses

### 400 Bad Request
```json
{
  "message": "Invalid request data",
  "errors": {
    "barcode": "required"
  }
}
```

### 404 Not Found
```json
{
  "message": "Scan not found"
}
```

### 500 Internal Server Error
```json
{
  "message": "Failed to create scan",
  "error": "database error details"
}
```

## Dependency Injection

The scan feature uses Uber FX for dependency injection:

```go
var Module = fx.Options(
    fx.Provide(
        repository.NewScanRepository,
        service.NewScanService,
        handler.NewScanHandler,
    ),
)
```

**Integration points:**
- Registered in `/internal/server/api/providers.go`
- Routes registered via `SetupRoutes()` in API server
- Services injected via FX constructor functions

## Testing

### Run All Scan Tests
```bash
go test -v ./internal/application/scan/...
```

### Test Coverage
```bash
go test -cover ./internal/application/scan/...
```

### Run Specific Test
```bash
go test -v -run TestScanHandler ./internal/application/scan/handler/...
```

## Usage Examples

### Create Scan with cURL
```bash
curl -X POST http://localhost:8080/api/v1/scans \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "barcode": "BC123456789",
    "timestamp": 1234567890000,
    "transaction_id": "trx-001",
    "pin": "123456",
    "photo": "data:image/jpeg;base64,...",
    "device": {
      "user_agent": "Mozilla/5.0",
      "platform": "Linux",
      "language": "en-US",
      "device_type": "mobile",
      "browser": "Chrome"
    },
    "photo_size": "45 KB"
  }'
```

### List Scans with Filters
```bash
curl -X GET "http://localhost:8080/api/v1/scans?page=1&page_size=10&status=pending" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Update Scan Status
```bash
curl -X PUT http://localhost:8080/api/v1/scans/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'
```

## Configuration

The scan module uses the following configuration (from `config.yaml`):

```yaml
server:
  host: localhost
  port: 8080

database:
  host: localhost
  port: 3306
  user: root
  password: password
  db_name: management_subscribe
```

## Logging

All operations are logged using structured logging with Zap:

```
[INFO]  Creating scan  {"user_id": 1, "barcode": "BC123456789"}
[INFO]  Scan created successfully  {"scan_id": 1}
[ERROR] Failed to create scan  {"error": "database error", "user_id": 1}
```

## Relationships

- **One-to-Many**: One User has many Scans
- **User Context**: Scans are tied to authenticated user via middleware
- **Soft Deletes**: Deleted scans are marked but not removed from database

## Best Practices

1. **Always include authorization headers** when calling API
2. **Validate photo size** before uploading (recommend max 5MB)
3. **Use transaction_id** to prevent duplicate scans
4. **Monitor status changes** for completed/failed scans
5. **Implement retry logic** for failed scans in client
6. **Clean up old scans** periodically (consider archiving)
7. **Log all scan operations** for audit trail

## Troubleshooting

### Scan not found
- Verify scan ID exists
- Check if scan was soft-deleted
- Confirm user has access to scan

### Invalid status
- Valid statuses: `pending`, `completed`, `failed`
- Check DTO binding tags
- Review error message

### Database connection
- Verify database credentials
- Check database is running
- Review connection pool settings

## Future Enhancements

1. **Batch operations** - Create multiple scans at once
2. **Export functionality** - Export scan data to CSV/PDF
3. **Webhook notifications** - Notify when scan status changes
4. **Analytics** - Track scan statistics and trends
5. **QR code support** - Extend to QR codes
6. **Photo verification** - ML-based photo validation
7. **Rate limiting** - Prevent scan spam
8. **Async processing** - Background job for heavy operations

## Files Structure

```
internal/application/scan/
├── dto/
│   └── scan.dto.go                    # Request/response DTOs
├── entity/
│   └── scan.entity.go                 # Domain entity
├── repository/
│   └── scan.repo.go                   # Data access layer
├── service/
│   └── scan.service.go                # Business logic
├── handler/
│   └── scan.handler.go                # HTTP handlers
├── module.go                          # FX dependency module
└── README.md                          # This file
```

## Support

For questions or issues:
1. Check API endpoint documentation
2. Review error logs
3. Verify database schema
4. Test with provided cURL examples
5. Contact development team
