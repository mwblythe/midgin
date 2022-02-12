# midgin

An adapter to use standard `net/http` middleware in [Gin](https://github.com/gin-gonic/gin).

## Overview

Gin is a very capable web framework, but it does not directly support standard `net/http` middleware. The `midgin` adapter makes it possible. This type of middleware has the following signature:

```go
func(next http.Handler) http.Handler
```

## Usage

```go
import (
  "github.com/gin-gonic/gin"
  "github.com/mwblythe/midgin"
  "github.com/rs/cors"
)

// use standard CORS middleware, for example
r := gin.Default()
r.Use(midgin.Adapt(cors.Default().Handler)
```

## 