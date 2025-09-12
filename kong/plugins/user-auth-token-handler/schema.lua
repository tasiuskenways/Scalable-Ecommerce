return {
  name = "user-auth-token-handler",
  fields = {
    { config = {
        type = "record",
        fields = {
          { public_key = { type = "string", default = "hmRkbgqWqgWrlYgDZmdslzQeKPoFQsirseqwXk5_EQ4" }, },
          { redis_host = { type = "string", default = "user-redis" }, },
          { redis_port = { type = "number", default = 6379 }, },
          { redis_password = { type = "string", required = false }, },
        },
      },
    },
  },
}
