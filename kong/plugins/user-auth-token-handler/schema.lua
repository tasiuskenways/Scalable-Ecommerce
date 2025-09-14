return {
  name = "user-auth-token-handler",
  fields = {
    { config = {
        type = "record",
        fields = {
          { public_key = { type = "string", default = "hmRkbgqWqgWrlYgDZmdslzQeKPoFQsirseqwXk5_EQ4" } },
          { redis_host = { type = "string", default = "user-redis" } },
          { redis_port = { type = "number", default = 6379 } },
          { redis_password = { type = "string", required = false } },
          { user_service_url = { type = "string", default = "http://user-service:3003" } },
          { timeout = { type = "number", default = 3000 } },
          -- RBAC Configuration
          { required_roles = { type = "array", elements = { type = "string" }, required = false } },
          { required_permissions = { type = "array", elements = { type = "string" }, required = false } },
          { owner_param = { type = "string", required = false } }, -- e.g., "userId"
          { allow_public = { type = "boolean", default = false } },
        },
      },
    },
  },
}