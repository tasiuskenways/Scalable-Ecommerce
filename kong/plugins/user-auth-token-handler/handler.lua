local jwt = require "resty.jwt"
local cjson = require "cjson.safe"
local redis = require "resty.redis"

local JwtBlacklistHandler = {
  PRIORITY = 2000,
  VERSION = "1.0",
}

function JwtBlacklistHandler:access(conf)
  local auth_header = kong.request.get_header("authorization")
  if not auth_header or not auth_header:find("Bearer ") then
    return kong.response.exit(401, { message = "Missing or invalid token" })
  end

  local token = auth_header:sub(8) -- remove "Bearer "
  
  -- verify JWT
  local jwt_obj = jwt:verify(conf.public_key, token)
  if not jwt_obj["verified"] then
    return kong.response.exit(401, { message = "Unauthorized" })
  end

  -- connect to Redis
  local red = redis:new()
  red:set_timeout(1000)
  local ok, err = red:connect(conf.redis_host, conf.redis_port)
  if not ok then
    kong.log.err("Redis connection failed: ", err)
    return kong.response.exit(500, { message = "Redis unavailable" })
  end

  if conf.redis_password and conf.redis_password ~= "" then
    local auth_ok, auth_err = red:auth(conf.redis_password)
    if not auth_ok then
      kong.log.err("Redis auth failed: ", auth_err)
      return kong.response.exit(500, { message = "Redis auth failed" })
    end
  end

  -- check blacklist (using jti or raw token)
  local jti = jwt_obj.payload.jti or token
  local res, err = red:get("blacklist:" .. jti)

  if res and res ~= ngx.null then
    return kong.response.exit(401, { message = "Token revoked" })
  end

  -- inject userId for downstream services
  kong.service.request.set_header("X-User-Id", jwt_obj.payload.sub)
end

return JwtBlacklistHandler