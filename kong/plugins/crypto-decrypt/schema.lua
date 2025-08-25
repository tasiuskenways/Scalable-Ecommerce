local typedefs = require "kong.db.schema.typedefs"

return {
  name = "crypto-decrypt",
  fields = {
    { consumer = typedefs.no_consumer },       -- allow plugin without consumer
    { protocols = typedefs.protocols_http },   -- only works for HTTP/HTTPS
    { config = {
        type = "record",
        fields = {}  -- no config for now
      }
    }
  }
}
