-- kong/plugins/user-auth-token-handler/handler.lua (updated)
local jwt = require "resty.jwt"
local redis = require "resty.redis"
local http = require "resty.http"
local cjson = require "cjson"

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
  local res, _ = red:get("blacklist:" .. jti)

  if res and res ~= ngx.null then
    return kong.response.exit(401, { message = "Token revoked" })
  end

  local user_id = jwt_obj.payload.user_id
  local user_email = jwt_obj.payload.email

  -- Get user roles and permissions from cache or user service
  local user_roles, user_permissions = get_user_roles_and_permissions(red, conf, user_id)

  -- Check access control if configured
  if conf.required_roles or conf.required_permissions or conf.owner_param then
    local access_granted = check_access_control(conf, user_id, user_roles, user_permissions)
    if not access_granted then
      return kong.response.exit(403, { 
        message = "Access denied",
        required_roles = conf.required_roles,
        required_permissions = conf.required_permissions
      })
    end
  end

  -- inject headers for downstream services
  kong.service.request.set_header("X-User-Id", user_id)
  kong.service.request.set_header("X-User-Email", user_email)
  kong.service.request.set_header("X-User-Roles", table.concat(user_roles or {}, ","))
  kong.service.request.set_header("X-User-Permissions", table.concat(user_permissions or {}, ","))
end

-- Function to get user roles and permissions
function get_user_roles_and_permissions(redis_client, conf, user_id)
  -- Try to get from Redis cache first
  local cache_key = "user_rbac:" .. user_id
  local cached_data, _ = redis_client:get(cache_key)
  
  if cached_data and cached_data ~= ngx.null then
    local ok, data = pcall(cjson.decode, cached_data)
    if ok and data.roles and data.permissions then
      kong.log.debug("[auth-token-handler] Using cached RBAC data for user: ", user_id)
      return data.roles, data.permissions
    end
  end

  -- If not in cache, fetch from user service
  kong.log.debug("[auth-token-handler] Fetching RBAC data from user service for user: ", user_id)
  
  local httpc = http.new()
  local res, err = httpc:request_uri(conf.user_service_url .. "/api/internal/users/" .. user_id .. "/rbac", {
    method = "GET",
    headers = {
      ["Content-Type"] = "application/json",
      ["X-Internal-Service"] = "kong-auth"
    },
    timeout = conf.timeout or 3000
  })

  if not res or res.status ~= 200 then
    kong.log.err("[auth-token-handler] Failed to fetch user RBAC data: ", err or res.status)
    return {}, {}
  end

  local ok, response_data = pcall(cjson.decode, res.body)
  if not ok or not response_data.success then
    kong.log.err("[auth-token-handler] Invalid RBAC response from user service")
    return {}, {}
  end

  local roles = response_data.data.roles or {}
  local permissions = response_data.data.permissions or {}

  -- Cache the result for 5 minutes
  local cache_data = cjson.encode({
    roles = roles,
    permissions = permissions
  })
  redis_client:setex(cache_key, 300, cache_data)

  return roles, permissions
end

-- Function to check access control
function check_access_control(conf, user_id, user_roles, user_permissions)
  local access_granted = false

  -- Check if public access is allowed
  if conf.allow_public then
    return true
  end

  -- Check required roles
  if conf.required_roles and #conf.required_roles > 0 then
    for _, required_role in ipairs(conf.required_roles) do
      for _, user_role in ipairs(user_roles or {}) do
        if user_role == required_role then
          kong.log.debug("[auth-token-handler] Access granted by role: ", required_role)
          return true
        end
      end
    end
  end

  -- Check required permissions
  if conf.required_permissions and #conf.required_permissions > 0 then
    for _, required_perm in ipairs(conf.required_permissions) do
      for _, user_perm in ipairs(user_permissions or {}) do
        if user_perm == required_perm then
          kong.log.debug("[auth-token-handler] Access granted by permission: ", required_perm)
          return true
        end
      end
    end
  end

  -- Check owner access
  if conf.owner_param then
    local path = kong.request.get_path()
    local resource_owner_id = path:match("/" .. conf.owner_param .. "/([^/]+)")
    if resource_owner_id and resource_owner_id == user_id then
      kong.log.debug("[auth-token-handler] Access granted as resource owner")
      return true
    end
  end

  return false
end

return JwtBlacklistHandler