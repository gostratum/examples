# Database Migrations

This directory contains versioned SQL migrations for the Order Service database.

## Structure

Each migration consists of two files:
- `{version}_{description}.up.sql` - Applies the migration
- `{version}_{description}.down.sql` - Rolls back the migration

## Current Migrations

| Version | Description | Files |
|---------|-------------|-------|
| 000001 | Create users table | `000001_create_users_table.{up,down}.sql` |
| 000002 | Create orders table | `000002_create_orders_table.{up,down}.sql` |
| 000003 | Create indexes | `000003_create_indexes.{up,down}.sql` |

## Adding New Migrations

1. **Determine the next version number** (e.g., 000004)

2. **Create the migration files**:
   ```bash
   touch migrations/000004_add_user_status.up.sql
   touch migrations/000004_add_user_status.down.sql
   ```

3. **Write the UP migration** - What to apply:
   ```sql
   -- migrations/000004_add_user_status.up.sql
   ALTER TABLE users ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active';
   CREATE INDEX idx_users_status ON users(status);
   ```

4. **Write the DOWN migration** - How to rollback:
   ```sql
   -- migrations/000004_add_user_status.down.sql
   DROP INDEX IF EXISTS idx_users_status;
   ALTER TABLE users DROP COLUMN status;
   ```

5. **Test the migration**:
   ```bash
   # Apply
   make migrate
   
   # Verify version
   make migrate-version
   
   # Test rollback
   make migrate-down STEPS=1
   
   # Re-apply
   make migrate
   ```

## Best Practices

### DO ✅

- **Use sequential version numbers** (000001, 000002, etc.)
- **Write descriptive migration names** (`add_user_status` not `update_users`)
- **Make migrations idempotent** when possible (use `IF NOT EXISTS`, `IF EXISTS`)
- **Test both up and down** migrations before committing
- **Keep migrations small** - one logical change per migration
- **Add comments** to explain complex SQL
- **Review before applying** in production

### DON'T ❌

- **Skip version numbers** - keep them sequential
- **Modify existing migrations** that have been applied
- **Use non-deterministic functions** (like `NOW()`) in migrations
- **Drop columns** in production without a deprecation plan
- **Forget the down migration** - always provide a rollback path
- **Mix DDL and DML** - separate schema changes from data changes

## Migration Naming Convention

```
{version}_{action}_{table}_{description}.{up|down}.sql
```

Examples:
- `000001_create_users_table.up.sql`
- `000002_add_orders_status_index.up.sql`
- `000003_alter_users_add_email_verification.up.sql`
- `000004_drop_deprecated_orders_fields.up.sql`

## Running Migrations

See [../MIGRATIONS.md](../MIGRATIONS.md) for complete documentation on running migrations.

Quick reference:
```bash
# Apply all pending
make migrate

# Rollback last
make migrate-down

# Check version
make migrate-version
```

## Troubleshooting

### Dirty State

If a migration fails partway through:

```bash
# Check current state
make migrate-version

# Force to last good version (e.g., 2)
make migrate-force VERSION=2

# Fix the migration file
# Then re-run
make migrate
```

### Testing Migrations

Always test in a development environment first:

```bash
# 1. Create a backup
pg_dump -h localhost -U postgres orders > backup.sql

# 2. Run migration
make migrate

# 3. Test your application
make test

# 4. Test rollback
make migrate-down

# 5. Verify rollback worked
make test

# 6. Re-apply for final test
make migrate
```

## Version Control

- ✅ Commit migration files to git
- ✅ Include in code reviews
- ✅ Tag releases with migration versions
- ❌ Never modify migrations that have been deployed
- ❌ Never delete migration files
