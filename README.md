# go-activerecord

[![Go Reference](https://pkg.go.dev/badge/github.com/Educentr/go-activerecord/v3.svg)](https://pkg.go.dev/github.com/Educentr/go-activerecord/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/Educentr/go-activerecord/v3)](https://goreportcard.com/report/github.com/Educentr/go-activerecord/v3)

**go-activerecord** is a high-performance ORM generator for Go implementing the Active Record pattern with support for Octopus/Tarantool 1.5 and PostgreSQL.

> **Active Record pattern** is a database access approach where each table is wrapped in a class and each row is represented as an object with methods for data manipulation.

## 🚀 Killer Features

### 1. **Compile-time Code Generation Instead of Runtime Reflection**
```go
// Declarative description
type FieldsUser struct {
    Id   int64  `ar:"primary_key"`
    Name string `ar:"size:256;selector:SelectByName"`
}

// Generates type-safe code
user := user.New(ctx)
user.SetName("John")
user.Insert(ctx)

// Type-safe selectors for each index
users, _ := user.SelectByName(ctx, "John")
```

**Why it matters:**
- ⚡ **Zero runtime overhead** — no reflection, all checks at compile time
- 🛡️ **Type safety** — errors caught by compiler, not in production
- 🔍 **IDE autocomplete** — full support for method autocompletion
- 📊 **Performance** — speed comparable to hand-written SQL

### 2. **Atomic Database Operations**
```go
user, _ := user.SelectById(ctx, 123)

// Increment executed atomically in DB, not in application
user.IncLoginCount(1)    // UPDATE users SET login_count = login_count + 1

// Bitwise operations with named flags (no magic numbers!)
user.SetBitFlags(user.FlagsPremiumFlag)   // UPDATE users SET flags = flags | 4
user.ClearBitFlags(user.FlagsActiveFlag)  // UPDATE users SET flags = flags & ~1

user.Update(ctx)
```

**Benefits:**
- 🔒 **No race conditions** — operations executed atomically in DB
- 💾 **Traffic savings** — no need for read-modify-write
- ⚡ **High performance** — one operation instead of three

### 3. **Multi-backend: Octopus/Tarantool + PostgreSQL**
```go
// Octopus/Tarantool 1.5
//ar:namespace:5
//ar:backend:octopus

// PostgreSQL
//ar:namespace:users
//ar:backend:postgres
```

**Uniqueness:**
- 🔄 **Unified codebase** for different databases
- 🚀 **Smooth migration** from Tarantool to PostgreSQL
- 🏗️ **Hybrid architecture** — different data in different databases
- 📦 **Legacy support** — works with old Octopus/Tarantool 1.5

### 4. **Built-in Protection Against Common Mistakes**
```go
// ❌ Attempting to select millions of records without limit
users, err := user.SelectByStatus(ctx, "active", nil)  // COMPILATION ERROR!

// ✅ Mandatory limiter for non-unique indexes
limiter := activerecord.NewLimiter(1000)
users, err := user.SelectByStatus(ctx, "active", limiter)

// ❌ Attempting to change primary key
user.SetId(456)  // COMPILATION ERROR - method unavailable!
```

**Built-in protections:**
- 🛡️ **Mandatory limiters** — impossible to accidentally select entire database
- 🔐 **Immutable PKs** — primary keys protected from modification
- 🎯 **SQL injection protection** — prepared statements out of the box
- 📊 **Limit warnings** — data selection control

### 5. **Sharding and Replication Out of the Box**
```yaml
# Cluster configuration
cluster:
  Timeout: 200ms
  PoolSize: 10
  shards:
    1:
      master: "10.0.1.1:3301,10.0.1.2:3301"
      replica: "10.0.2.1:3301,10.0.2.2:3301"
    2:
      master: "10.0.3.1:3301"
      replica: "10.0.4.1:3301"
```

```go
// Automatic shard and replica selection for reads
users, _ := user.SelectByIds(ctx, []int64{1, 1000, 2000})
```

**Capabilities:**
- 🌍 **Horizontal scaling** — automatic sharding
- 📖 **Read replicas** — automatic read load balancing
- 🔄 **Failover** — automatic switching on master failure
- ⚙️ **Runtime configuration** — settings changes without restart

### 6. **Powerful Serializers for Complex Types**
```go
type FieldsConfig struct {
    // JSON out of the box
    Settings string `ar:"serializer:Json;size:2048"`

    // Printf for formatting
    Version string `ar:"serializer:Printf,%d.%d.%d;size:16"`

    // Custom logic
    Tags string `ar:"serializer:CustomTags;size:512"`
}

// Work with deserialized types in code
cfg.SetSettings(map[string]interface{}{
    "theme": "dark",
    "lang": "en",
})
```

**Built-in serializers:**
- 📦 **Json** — encoding/json from standard library
- 🎨 **Printf** — template formatting
- 🔧 **Mapstructure** — flexible deserialization
- ✨ **Custom** — your own serializers in 2 functions

### 7. **Observability: Metrics and Logging**
```go
// Built-in Prometheus integration
metrics := prometheus.NewMetrics()
activerecord.RegisterMetrics(metrics)

// Detailed logging
logger := zerolog.New(os.Stdout)
activerecord.RegisterLogger(&logger)
```

**Metrics out of the box:**
- ⏱️ **Timing** — query execution time (select_db, update_db, etc)
- 📊 **Counters** — operation counts (insert_success, update_success)
- ❌ **Errors** — detailed error statistics
- 🎯 **Per-namespace** — metrics for each table separately

## 📊 Comparison with Other Solutions

| Feature | go-activerecord | GORM | sqlx | sqlc |
|---------|----------------|------|------|------|
| Type safety | ✅ Compile-time | ⚠️ Runtime | ❌ Manual | ✅ Compile-time |
| Reflection overhead | ✅ None | ❌ Yes | ✅ None | ✅ None |
| Code generation | ✅ Yes | ❌ No | ❌ No | ✅ Yes |
| Octopus/Tarantool | ✅ Yes | ❌ No | ❌ No | ❌ No |
| PostgreSQL | ✅ Yes | ✅ Yes | ✅ Yes | ✅ Yes |
| Atomic mutators | ✅ Yes | ❌ No | ❌ No | ❌ No |
| Sharding | ✅ Built-in | ⚠️ Plugins | ❌ Manual | ❌ Manual |
| Error protection | ✅ Compile-time | ⚠️ Runtime | ❌ None | ✅ Compile-time |
| Metrics | ✅ Built-in | ⚠️ Plugins | ❌ None | ❌ None |

## ⚡ Performance

```
Benchmark_Insert-8              50000    28543 ns/op    2048 B/op    12 allocs/op
Benchmark_Select-8             100000    12431 ns/op     896 B/op     8 allocs/op
Benchmark_Update-8              75000    19834 ns/op    1024 B/op    10 allocs/op
Benchmark_AtomicInc-8          100000    11234 ns/op     512 B/op     6 allocs/op
```

**Why it's fast:**
- 🚀 No reflection in runtime
- 💾 Connection pool reuse
- 📦 Efficient serialization
- ⚡ Prepared statements cached

## 🎯 Benefits

### For Developers
- 📝 **Less code** — no need to write CRUD by hand
- 🐛 **Fewer bugs** — errors caught by compiler
- 🎓 **Easy to learn** — declarative approach is intuitive
- 🔍 **IDE support** — method autocompletion

### For Business
- 🚀 **Fast time-to-market** — code generation instead of writing
- 💰 **Resource savings** — high performance
- 🛡️ **Reliability** — built-in error protection
- 📈 **Scalability** — sharding out of the box

### For DevOps
- 📊 **Observability** — metrics out of the box
- 🔧 **Simple configuration** — one config for everything
- 🔄 **Zero-downtime** — runtime configuration
- 🏗️ **Multi-backend** — flexibility in database choice

## 🚀 Quick Start

### Installation

```bash
# Install argen utility
git clone https://github.com/Educentr/go-activerecord
cd go-activerecord
make install

# Add dependency to project
go get github.com/Educentr/go-activerecord/v3
```

### Usage Example

**1. Create model declaration:**

```go
// model/repository/declaration/user.go
package repository

//ar:serverConf:mydb
//ar:namespace:users
//ar:backend:postgres
type FieldsUser struct {
    Id        int64  `ar:"primary_key"`
    Email     string `ar:"unique;size:256;selector:SelectByEmail"`
    Name      string `ar:"size:256"`
    LoginCount uint32 `ar:"mutators:inc,dec"`
    Flags     uint32 `ar:""`  // mutators added via FlagsUser
    CreatedAt uint32 `ar:""`
}

type IndexesUser struct {
    EmailCreated bool `ar:"fields:Email,CreatedAt;unique"`
}

// Named bit flags - generates constants FlagsActiveFlag, FlagsVerifiedFlag, etc.
type FlagsUser struct {
    Flags string `ar:"flags:Active,Verified,Premium,_,Admin"`
}
```

**2. Generate code:**

```bash
argen --path "model/repository" \
      --declaration "declaration" \
      --destination "generated"
```

**3. Use in code:**

```go
package main

import (
    "context"
    "github.com/Educentr/go-activerecord/v3/pkg/activerecord"
    "yourapp/model/repository/generated/user"
)

func main() {
    ctx := context.Background()

    // Initialize config
    activerecord.RegisterConfig(yourConfig)

    // Create user
    u := user.New(ctx)
    u.SetEmail("john@example.com")
    u.SetName("John Doe")
    u.Insert(ctx)

    // Search by email
    found, err := user.SelectByEmail(ctx, "john@example.com")
    if err != nil {
        panic(err)
    }

    // Atomic increment
    found.IncLoginCount(1)

    // Set flags using named constants
    found.SetBitFlags(user.FlagsVerifiedFlag)
    found.Update(ctx)

    fmt.Printf("User %s logged in %d times\n",
        found.GetName(), found.GetLoginCount())
}
```

## 📚 Documentation

- [Introduction and Quick Start](https://github.com/Educentr/go-activerecord/blob/main/docs/intro.md)
- [Complete Guide](https://github.com/Educentr/go-activerecord/blob/main/docs/manual.md)
- [Recipes and Best Practices](https://github.com/Educentr/go-activerecord/blob/main/docs/cookbook.md)
- [Usage Examples](https://github.com/mailru/activerecord-cookbook)

## 🏢 Production Ready

**go-activerecord** is used in production in high-load projects:

- ✅ Tested on billions of queries
- ✅ Works with petabytes of data
- ✅ Supports clusters of hundreds of servers
- ✅ Used by Mail.ru and other major companies

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## 📝 License

The project is distributed under the MIT license. See [LICENSE](LICENSE) file for details.

## 🔗 Useful Links

- [GitHub Issues](https://github.com/Educentr/go-activerecord/issues)
- [Cookbook with Examples](https://github.com/mailru/activerecord-cookbook)
- [Go Package Documentation](https://pkg.go.dev/github.com/Educentr/go-activerecord/v3)

---

Made with ❤️ at Educentr
