local http = require "resty.http"
local cjson = require "cjson"

local CryptoDecryptHandler = {
  PRIORITY = 1000,
  VERSION = "1.0",
}

function CryptoDecryptHandler:access(conf)
  kong.log.debug("[crypto-decrypt] running access phase")

  ngx.req.read_body()
  local body = ngx.req.get_body_data()

  if not body then
    kong.log.err("[crypto-decrypt] request body is empty but encryption required")
    return kong.response.exit(400, { message = "Encrypted request body required" })
  end

  local ok, json = pcall(cjson.decode, body)
  if not ok or not json["data"] then
    kong.log.err("[crypto-decrypt] missing 'data' field in request body")
    return kong.response.exit(400, { message = "Bad request" })
  end

  -- Call crypto-service
  local httpc = http.new()
  local res, err = httpc:request_uri("http://crypto-service:3002/api/decrypt", {
    method = "POST",
    body = body,
    headers = {
      ["Content-Type"] = "application/json"
    }
  })

  kong.log.debug("[crypto-decrypt] crypto-service response: ", res)

  if not res then
    kong.log.err("[crypto-decrypt] crypto-service request failed: ", err)
    return kong.response.exit(500, { message = "Decryption service unavailable" })
  end

  if res.status ~= 200 then
    kong.log.err("[crypto-decrypt] crypto-service error: ", res.body)
    return kong.response.exit(500, { message = "Decryption failed" })
  end

  if not res.body then
    kong.log.err("[crypto-decrypt] crypto-service returned empty body")
    return kong.response.exit(500, { message = "Invalid decryption response" })
  end

  -- Parse decrypted body
  local ok2, decrypted = pcall(cjson.decode, res.body)
  if not ok2 then
    kong.log.err("[crypto-decrypt] failed to decode decrypted JSON")
    return kong.response.exit(500, { message = "Invalid decrypted response" })
  end

  local ok3, decrypted_json = pcall(cjson.encode, decrypted.data)
  if ok3 then
    kong.log.debug("[crypto-decrypt] decrypted body JSON: ", decrypted_json)
  else
    kong.log.debug("[crypto-decrypt] decrypted body (raw table)")
  end

  -- Replace request body with proper JSON
  local final_body = decrypted.data
  if type(final_body) == "table" then
    final_body = cjson.encode(final_body)
  end
  ngx.req.set_body_data(final_body)

  kong.log.debug("[crypto-decrypt] final body injected: ", final_body)

  kong.log.debug("[crypto-decrypt] successfully decrypted body")
end

return CryptoDecryptHandler
